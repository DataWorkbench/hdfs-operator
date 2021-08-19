package datanode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/controllers/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DatanodeScripts            = "datanode-scripts"
	LivenessAndReadinessConfigKey = "check-status.sh"
)

func BuildConfigMap(hdfs v1.HDFS)  corev1.ConfigMap{

	configmap := types.NamespacedName{Namespace: hdfs.Namespace, Name: com.GetName(hdfs.Name,DatanodeScripts)}

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmap.Name,
			Namespace: configmap.Namespace,
			Labels:    com.NewLabels(configmap),
		},
		Data: map[string]string{
			LivenessAndReadinessConfigKey: LivenessAndReadinessScript,
		},
	}
}


const LivenessAndReadinessScript = `#!/usr/bin/env bash
    # Exit on error. Append "|| true" if you expect an error.
    set -o errexit
    # Exit on error inside any functions or subshells.
    set -o errtrace
    # Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
    set -o nounset
    set -o pipefail
    # Turn on traces, useful while debugging.
    set -o xtrace

    # Check if datanode registered with the namenode and got non-null cluster ID.
    _PORTS="50075 1006"
    _URL_PATH="jmx?qry=Hadoop:service=DataNode,name=DataNodeInfo"
    _CLUSTER_ID=""
    for _PORT in $_PORTS; do
      _CLUSTER_ID+=$(curl -s http://localhost:${_PORT}/$_URL_PATH |  \
          grep ClusterId) || true
    done
    echo $_CLUSTER_ID | grep -q -v null
`
