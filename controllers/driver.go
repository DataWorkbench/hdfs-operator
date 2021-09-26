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

	//// State holds the accumulated state during the reconcile loop
	//ReconcileState *reconcile.State
	//// Observers that observe es clusters state.
	//Observers *observer.Manager
}

func (d *DefaultDriver) Reconcile(ctx context.Context) *Results {

	results := &Results{ctx: ctx}
	//ObservedStateResolver Monitor the status of components
	//d.Observers.ObservedStateResolver(){}

	// reconcile StatefulSets and nodes configuration
	_ = d.reconcileNodeSpecs(ctx)
	//results = results.WithResults(res)
	//d.ReconcileState.UpdateHdfsState(*resourcesState, observedState)

	return results
}

func (d *DefaultDriver) reconcileNodeSpecs(ctx context.Context) *Results {
	results := &Results{}
	////step1  Parsing customer kind HDFS
	expectedResources, err := BuildExpectedResources(d.Hdfs)
	if err != nil {
		return results.WithError(err)
	}
	//step2 apply expected k8s kind
	upscaleResults, err := HandleUpscaleAndSpecChanges(d.Client, d.Hdfs, expectedResources)

	if upscaleResults.Requeue {
		//return results.WithResult(defaultRequeue)
		return results
	}
	return results
}