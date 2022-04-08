package namenode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	NamenodeScripts = "namenode-scripts"
	FormatConfigKey = "format-and-run.sh"
)

func BuildConfigMap(hdfs v1.HDFS) corev1.ConfigMap {

	configmap := types.NamespacedName{Namespace: hdfs.Namespace, Name: com.GetName(hdfs.Name, NamenodeScripts)}

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
			FormatConfigKey: getContainerArgsScript(hdfs.Spec.Version),
		},
	}
}

func getContainerArgsScript(version string) string {

	script := `#!/usr/bin/env bash
    set -o errexit
    set -o errtrace
    set -o nounset
    set -o pipefail
    set -o xtrace
    _HDFS_BIN=`

	hadoopHome := "$HADOOP_PREFIX"

	containerArgsScript := `/bin/hdfs
    _METADATA_DIR=/hadoop/dfs/name/current
    if [[ "$MY_POD" = "$NAMENODE_POD_0" ]]; then
      if [[ ! -d $_METADATA_DIR ]]; then
          $_HDFS_BIN --config $HADOOP_CONF_DIR namenode -format  \
              -nonInteractive hdfs-k8s ||
              (rm -rf $_METADATA_DIR; exit 1)
      fi
      _ZKFC_FORMATTED=/hadoop/dfs/name/current/.hdfs-k8s-zkfc-formatted
      if [[ ! -f $_ZKFC_FORMATTED ]]; then
        _OUT=$($_HDFS_BIN --config $HADOOP_CONF_DIR zkfc -formatZK -nonInteractive 2>&1)
        (echo $_OUT | grep -q "FATAL") && exit 1
        touch $_ZKFC_FORMATTED
      fi
    elif [[ "$MY_POD" = "$NAMENODE_POD_1" ]]; then
      if [[ ! -d $_METADATA_DIR ]]; then
        sleep 1m
        $_HDFS_BIN --config $HADOOP_CONF_DIR namenode -bootstrapStandby  \
            -nonInteractive ||  \
            (rm -rf $_METADATA_DIR; exit 1)
      fi
    fi
    `

	if version[0:1] == "3" {
		hadoopHome = "$HADOOP_HOME"
		return  script+hadoopHome+containerArgsScript + `
            nohup $_HDFS_BIN --config $HADOOP_CONF_DIR zkfc &
            $_HDFS_BIN --config $HADOOP_CONF_DIR namenode  `
	}
	//return  script+hadoopHome+containerArgsScript +`
    //         $HADOOP_PREFIX/sbin/hadoop-daemon.sh --config $HADOOP_CONF_DIR start zkfc
    //         $_HDFS_BIN --config $HADOOP_CONF_DIR namenode  `
	return  script+hadoopHome+containerArgsScript +`
             nohup $_HDFS_BIN --config $HADOOP_CONF_DIR zkfc &
             $_HDFS_BIN --config $HADOOP_CONF_DIR namenode  `
}