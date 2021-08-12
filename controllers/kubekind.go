package controllers

import (
	"fmt"
	hdfsv1 "github.com/dataworkbench/hdfs-operator/api/v1"
	"k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpscaleResults struct {
	ActualStatefulSet v1.StatefulSet
	Requeue            bool
}

func HandleUpscaleAndSpecChanges( c client.Client,hdfs hdfsv1.HDFS,res Resources) (UpscaleResults, error) {
	results := UpscaleResults{}

	   //kind: ConfigMap  name: my-hdfs-config
	//ReconcileConfig(ctx.k8sClient,res.Config)
	  //kind: Service  name: my-hdfs-config
	//ReconcileService( ctx.k8sClient, &res.HeadlessService);

	_/*reconciled*/, err := ReconcileStatefulSet(c, hdfs, res.StatefulSet)

	if err != nil {
		return results, fmt.Errorf("reconcile StatefulSet: %w", err)
	}
	// update actual with the reconciled ones for next steps to work with up-to-date information
	//results.ActualStatefulSet = actualStatefulSets.WithStatefulSet(reconciled)
	return results, nil

}

// ReconcileStatefulSet creates or updates the statefulset kind
func ReconcileStatefulSet( c client.Client,hdfs hdfsv1.HDFS, expected v1.StatefulSet,) (v1.StatefulSet, error) {
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
