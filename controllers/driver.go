package controllers

import (
	"context"
	"github.com/dataworkbench/hdfs-operator/api/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DefaultDriver struct {
	// Hdfs is the HDFS resource to reconcile
	Hdfs v1.HDFS
	// Client is used to access the Kubernetes API.
	Client   client.Client

	Recorder record.EventRecorder

	//// LicenseChecker is used for some features to check if an appropriate license is setup
	//LicenseChecker commonlicense.Checker
	//// State holds the accumulated state during the reconcile loop
	//ReconcileState *reconcile.State
	//// Observers that observe es clusters state.
	//Observers *observer.Manager
	//// DynamicWatches are handles to currently registered dynamic watches.
	//DynamicWatches watches.DynamicWatches
	//// Expectations control some expectations set on resources in the cache, in order to
	//// avoid doing certain operations if the cache hasn't seen an up-to-date resource yet.
	//Expectations *expectations.Expectations
}

func (d *DefaultDriver) Reconcile(ctx context.Context) *Results {

	results :=&Results{ctx: ctx}

	//ObservedStateResolver Monitor the status of components such as NN and DN.
	//For example, NN is not in activity status, or some DN and NN lose heartbeat
	//Then an event is sent to the relevant channel of the watch capability provided by k8s controller runtime
	//to trigger the operator to start reconciling process correction
	//d.Observers.ObservedStateResolver(){}

	// reconcile StatefulSets and nodes configuration
	_ = d.reconcileNodeSpecs(ctx)
	//results = results.WithResults(res)

	//d.ReconcileState.UpdateHdfsState(*resourcesState, observedState)

	return results
}

func (d *DefaultDriver) reconcileNodeSpecs(ctx context.Context) *Results {

	results :=&Results{}

	//step1  Parsing customer kind HDFS
	expectedResources, err := BuildExpectedResources(d.Client, d.Hdfs)
	if err != nil {
		return results.WithError(err)
	}

	//step2 apply expected k8s kind resources and scale up.
	upscaleResults, err := HandleUpscaleAndSpecChanges(d.Client, d.Hdfs, expectedResources )

	if upscaleResults.Requeue {
		//return results.WithResult(defaultRequeue)
		return results
	}
	return results

}




