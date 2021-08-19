package journalnode

import (
	v1 "github.com/dataworkbench/hdfs-operator/api/v1"
	com "github.com/dataworkbench/hdfs-operator/controllers/common"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DefaultStorageClassName = "journalnode-disks"

func BuildStatefulSet( hdfs v1.HDFS) (appsv1.StatefulSet, error) {
	statefulSetName := com.GetName(hdfs.Name, hdfs.Spec.Journalnode.Name)
	// ssetSelector is used to match the StatefulSet pods
	ssetSelector := com.NewStatefulSetLabels(com.ExtractNamespacedName(&hdfs),statefulSetName)
	// add default PVCs to the node spec if no user defined PVCs exist
	hdfs.Spec.Journalnode.VolumeClaimTemplates = com.AppendDefaultPVCs(hdfs.Spec.Namenode.VolumeClaimTemplates,"editdir",DefaultStorageClassName)
	// build pod template
	podTemplate, err := BuildPodTemplateSpec(hdfs)
	if err != nil {
		return appsv1.StatefulSet{}, err
	}

	sset := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hdfs.Namespace,
			Name:      statefulSetName,
			Labels:    ssetSelector,
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			RevisionHistoryLimit: nil,
			ServiceName: statefulSetName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ssetSelector,
			},
			Replicas:             &hdfs.Spec.Namenode.Replicas,
			VolumeClaimTemplates: hdfs.Spec.Journalnode.VolumeClaimTemplates,
			Template:             podTemplate,
		},
	}
	return sset, nil
}



