package controllers

import (
	"fmt"
	hdfsv1 "github.com/dataworkbench/hdfs-operator/api/v1"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpscaleResults struct {
	ActualStatefulSet v1.StatefulSet
	Requeue            bool
}

func HandleUpscaleAndSpecChanges( c client.Client,hdfs hdfsv1.HDFS,res HdfsResources) (UpscaleResults, error) {

	//results := UpscaleResults{}
	results, err :=HandleConfigChanges(c, hdfs, res.CommonConfig )
	if err != nil {
		return results, fmt.Errorf("reconcile hdfs StatefulSets: %w", err)
	}

	for _,r := range res.Nodes{
		results, err :=HandleNodeSpecChanges(c, hdfs, r )
		if err != nil {
			return results, fmt.Errorf("reconcile hdfs StatefulSets: %w", err)
		}
	}
	results, err =HandleDNSpecChanges(c, hdfs, res.Datanode )
	if err != nil {
		return results, fmt.Errorf("reconcile hdfs : %w", err)
	}
	return results, nil
}

func HandleConfigChanges( c client.Client,hdfs hdfsv1.HDFS,cfg corev1.ConfigMap) (UpscaleResults, error) {
	results := UpscaleResults{}

	_/*reconciled*/, err := ReconcileCommonConfig(c, cfg, &hdfs)
	if err != nil {
		return results, fmt.Errorf("reconcile  config: %w", err)
	}
	//results.ConfigMap =

	return results, nil
}

func HandleNodeSpecChanges( c client.Client,hdfs hdfsv1.HDFS,res NodeResources) (UpscaleResults, error) {

	results, err :=HandleConfigChanges(c, hdfs, res.Config )
	if err != nil {
		return results, fmt.Errorf("reconcile hdfs StatefulSets: %w", err)
	}

	if _,err := ReconcileService(c, &res.HeadlessService,&hdfs); err != nil {
		return results, fmt.Errorf("reconcile service: %w", err)  //kind: Service  name: my-hdfs-config
	}
	_/*reconciled*/, err = ReconcileStatefulSet(c, hdfs, res.StatefulSet)

	if err != nil {
		return results, fmt.Errorf("reconcile StatefulSet: %w", err)
	}
	// update actual with the reconciled ones for next steps to work with up-to-date information
	//results.ActualStatefulSet = actualStatefulSets.WithStatefulSet(reconciled)
	return results, nil
}

func HandleDNSpecChanges( c client.Client,hdfs hdfsv1.HDFS,res DataResources) (UpscaleResults, error) {

	results, err :=HandleConfigChanges(c, hdfs, res.Config )
	if err != nil {
		return results, fmt.Errorf("reconcile hdfs StatefulSets: %w", err)
	}

	_/*reconciled*/, err = ReconcileDaemonSet(c, hdfs, res.DaemonSet)
	if err != nil {
		return results, fmt.Errorf("reconcile DaemonSet: %w", err)
	}

	//results.ActualDaemonSet = actualDaemonSet.WithDaemonSet(reconciled)
	return results, nil
}

// ReconcileStatefulSet creates or updates the statefulset kind
func ReconcileStatefulSet(c client.Client,hdfs hdfsv1.HDFS, expected v1.StatefulSet,) (v1.StatefulSet, error) {
	//podTemplateValidator := newPodTemplateValidator(c, es, expected)

	//create kind instance
	var reconciled v1.StatefulSet
	err := ReconcileResource(Params{
		Client:     c,
		Owner:      &hdfs,
		Expected:   &expected,
		Reconciled: &reconciled,
	})

	return reconciled, err
}

// ReconcileDaemonSet creates or updates the DaemonSet kind
func ReconcileDaemonSet(c client.Client,hdfs hdfsv1.HDFS, expected v1.DaemonSet,) (v1.DaemonSet, error) {

	//create kind instance
	var reconciled v1.DaemonSet
	err := ReconcileResource(Params{
		Client:     c,
		Owner:      &hdfs,
		Expected:   &expected,
		Reconciled: &reconciled,
	})

	return reconciled, err
}

func ReconcileCommonConfig(c client.Client, expected corev1.ConfigMap, owner client.Object) (corev1.ConfigMap, error) {
	var reconciled corev1.ConfigMap
	if err := ReconcileResource(Params{
		Client:     c,
		Owner:      owner,
		Expected:   &expected,
		Reconciled: &reconciled,
	}); err != nil {
		return corev1.ConfigMap{}, err
	}
	return reconciled, nil
}

func ReconcileService(
	c client.Client,
	expected *corev1.Service,
	owner client.Object,
) (*corev1.Service, error) {

	reconciled := &corev1.Service{}
	err := ReconcileResource(Params{
		Client:     c,
		Owner:      owner,
		Expected:   expected,
		Reconciled: reconciled,
	})
	return reconciled, err
}

