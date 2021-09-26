package datanode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/common"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DNDataVolumeName      = "hdfs-data-0"
	DNDataVolumeMountPath = "/hadoop/dfs/data/0"
	DNDataHostPath        = "/mnt/hdfs/dn-data"

	DNScriptsVolumeName      = "dn-scripts"
	DNScriptsVolumeMountPath = "/dn-scripts"
)

var defaultOptional = true

func BuildDaemonSet(hdfs v1.HDFS) (appsv1.DaemonSet, error) {
	daemonSetName := com.GetName(hdfs.Name, hdfs.Spec.Datanode.Name)
	selector := com.NewLabels(com.ExtractNamespacedName(&hdfs))
	// build pod template
	podTemplate, err := BuildDataNodePod(hdfs, selector)
	if err != nil {
		return appsv1.DaemonSet{}, err
	}

	daemonSet := appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonSetName,
			Namespace: hdfs.Namespace,
			Labels:    selector,
			OwnerReferences: com.GetOwnerReference(hdfs),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Template: podTemplate,
		},
	}

	return daemonSet, nil
}
