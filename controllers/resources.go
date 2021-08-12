package controllers

import (
	"github.com/dataworkbench/hdfs-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resources struct {
	StatefulSet     appsv1.StatefulSet
	HeadlessService corev1.Service
	//Config          settings.CanonicalConfig
}


func BuildExpectedResources(c client.Client, hdfs v1.HDFS) (Resources, error) {

	// build stateful set and associated headless service
	statefulSet, err := BuildStatefulSet(c, hdfs)
	if err != nil {
		return Resources{}, err
	}
	headlessSvc := HeadlessService(c,&hdfs, statefulSet.Name)

	return Resources{
		StatefulSet:     statefulSet,
		HeadlessService: headlessSvc,
		//Config:          cfg,
	},nil

}

// HeadlessService returns a headless service for the given StatefulSet
func HeadlessService( c client.Client ,hdfs *v1.HDFS, ssetName string) corev1.Service {
	nsn := ExtractNamespacedName(hdfs)
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nsn.Namespace,
			Name:      ssetName,
			Labels:    NewStatefulSetLabels(nsn, ssetName),
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Selector:  NewStatefulSetLabels(nsn, ssetName),
			Ports: getDefaultServicePorts(),
			},
		}
}

func BuildStatefulSet(c client.Client, hdfs v1.HDFS) (appsv1.StatefulSet, error) {

	namenode := hdfs.Spec.Namenode
	statefulSetName := namenode.Name //hdfs.StatefulSetName(hdfs.Name, hdfs.Spec.Namenode.Name)

	// ssetSelector is used to match the StatefulSet pods
	ssetSelector := NewStatefulSetLabels(ExtractNamespacedName(&hdfs), statefulSetName)

	// build pod template
	podTemplate, err := BuildPodTemplateSpec(c, hdfs)
	if err != nil {
		return appsv1.StatefulSet{}, err
	}

	// build sset labels on top of the selector
	ssetLabels := make(map[string]string)
	for k, v := range ssetSelector {
		ssetLabels[k] = v
	}

	sset := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hdfs.Namespace,
			Name:      statefulSetName,
			Labels:    ssetLabels,
		},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			// use default revision history limit
			RevisionHistoryLimit: nil,
			ServiceName: statefulSetName, //matching the StatefulSet labels
			Selector: &metav1.LabelSelector{
				MatchLabels: ssetSelector,
			},

			Replicas:             &hdfs.Spec.Namenode.Replicas,
			VolumeClaimTemplates: namenode.VolumeClaimTemplates,
			Template:             podTemplate,
		},
	}

	return sset, nil
	
}


