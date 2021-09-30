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

func BuildStatefulSet(hdfs v1.HDFS) (appsv1.StatefulSet, error) {
	statefulSetName := com.GetName(hdfs.Name, hdfs.Spec.Datanode.Name)
	// ssetSelector is used to match the StatefulSet pods
	ssetSelector := com.NewStatefulSetLabels(com.ExtractNamespacedName(&hdfs), statefulSetName)

	hdfs.Spec.Datanode.VolumeClaimTemplates = com.AppendDefaultPVCs(hdfs.Spec.Datanode.VolumeClaimTemplates,
		DNDataVolumeName, hdfs.Spec.Datanode.StorageClass)

	// build pod template,associate PVCs to pod container
	podTemplate, err := BuildPodTemplateSpec(hdfs, ssetSelector)
	if err != nil {
		return appsv1.StatefulSet{}, err
	}

	sset := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hdfs.Namespace,
			Name:      statefulSetName,
			Labels:    ssetSelector,
			OwnerReferences: com.GetOwnerReference(hdfs),
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			RevisionHistoryLimit: nil,
			//ServiceName:          statefulSetName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ssetSelector,
			},
			Replicas:             &hdfs.Spec.Datanode.Replicas,
			VolumeClaimTemplates: hdfs.Spec.Datanode.VolumeClaimTemplates,
			Template:             podTemplate,
		},
	}
	return sset, nil
}
