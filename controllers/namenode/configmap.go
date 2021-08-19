package namenode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/controllers/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	NamenodeScripts     = "namenode-scripts"
	FormatConfigKey     = "format-and-run.sh"
)

func BuildConfigMap(hdfs v1.HDFS)  corev1.ConfigMap{

	configmap := types.NamespacedName{Namespace: hdfs.Namespace, Name: com.GetName(hdfs.Name,NamenodeScripts)}

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmap.Name,
			Namespace: configmap.Namespace,
			Labels:    com.NewLabels(configmap),
		},
		Data: map[string]string{
			FormatConfigKey: ContainerArgsScript,
		},
	}
}


const ContainerArgsScript = `#!/usr/bin/env bash
    # Exit on error. Append "|| true" if you expect an error.
    set -o errexit
    # Exit on error inside any functions or subshells.
    set -o errtrace
    # Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
    set -o nounset
    set -o pipefail
    # Turn on traces, useful while debugging.
    set -o xtrace

    _HDFS_BIN=$HADOOP_PREFIX/bin/hdfs
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
        # zkfc masks fatal exceptions and returns exit code 0
        (echo $_OUT | grep -q "FATAL") && exit 1
        touch $_ZKFC_FORMATTED
      fi
    elif [[ "$MY_POD" = "$NAMENODE_POD_1" ]]; then
      if [[ ! -d $_METADATA_DIR ]]; then
        $_HDFS_BIN --config $HADOOP_CONF_DIR namenode -bootstrapStandby  \
            -nonInteractive ||  \
            (rm -rf $_METADATA_DIR; exit 1)
      fi
    fi
    $HADOOP_PREFIX/sbin/hadoop-daemon.sh --config $HADOOP_CONF_DIR start zkfc
    $_HDFS_BIN --config $HADOOP_CONF_DIR namenode
`