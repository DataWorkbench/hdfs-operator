package datanode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

const (
	DatanodeScripts               = "datanode-scripts"
	LivenessAndReadinessConfigKey = "check-status.sh"
)

func BuildConfigMap(hdfs v1.HDFS) corev1.ConfigMap {

	configmap := types.NamespacedName{Namespace: hdfs.Namespace, Name: com.GetName(hdfs.Name, DatanodeScripts)}

	return corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmap.Name,
			Namespace: configmap.Namespace,
			Labels:    com.NewLabels(configmap),
			OwnerReferences: com.GetOwnerReference(hdfs),
		},
		Data: map[string]string{
			LivenessAndReadinessConfigKey: liveScript+ strconv.Itoa(com.DatanodeRpcPort)+readinessScript,
		},
	}
}

var liveScript = `#!/usr/bin/env bash 
     _PORTS=` +"\""
var readinessScript =	" 1006\""+
	`
     _URL_PATH="jmx?qry=Hadoop:service=DataNode,name=DataNodeInfo"
     _CLUSTER_ID=""
     for _PORT in $_PORTS; do
       _CLUSTER_ID+=$(curl -s http://localhost:${_PORT}/$_URL_PATH |  \
           grep ClusterId) || true
     done
     echo $_CLUSTER_ID | grep -q -v null`

