/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/dataworkbench/hdfs-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// HDFSReconciler reconciles a HDFS object
type HDFSReconciler struct {
	client.Client
	recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=qy.dataworkbench.com,resources=hdfs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=qy.dataworkbench.com,resources=hdfs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=qy.dataworkbench.com,resources=hdfs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HDFS object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *HDFSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)

	// Fetch the HDFS instance
	var hdfs v1.HDFS
	requeue, err := r.fetchHdfsKind(ctx, req, &hdfs)
	if err != nil || requeue {
		return reconcile.Result{}, nil //tracing.CaptureError(ctx, err)
	}

	state := NewState(hdfs)
	results := r.internalReconcile(ctx, hdfs, state)

	return results.WithError(err).Aggregate()
}

func (r *HDFSReconciler) fetchHdfsKind(ctx context.Context, request reconcile.Request, hdfs *v1.HDFS) (bool, error) {

	err := r.Client.Get(ctx, request.NamespacedName, hdfs) //FetchWithAssociations
	if err != nil {
		if errors.IsNotFound(err) { // Object not found,
			// Object not found, cleanup in-memory state. Children resources are garbage-collected either by
			// the operator (see `onDelete`), either by k8s through the ownerReference mechanism.
			//return true, r.onDelete(types.NamespacedName{
			//	Namespace: request.Namespace,
			//	Name:      request.Name,
			//})
			return true, err
		}
		return true, err // Error reading the object - requeue the request.
	}
	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HDFSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.HDFS{}).
		Complete(r)
}

func (r *HDFSReconciler) internalReconcile(ctx context.Context, hdfs v1.HDFS, state *State) *Results {

	results := &Results{ctx: ctx}

	if hdfs.IsMarkedForDeletion() {
		// resource will be deleted, nothing to reconcile
		return results.WithError(nil) //return results.WithError(r.onDelete(k8s.ExtractNamespacedName(&es)))
	}

	driver := DefaultDriver{
		Hdfs:   hdfs,
		Client: r.Client,
		//ReconcileState:     state,
		Recorder: r.recorder,
	}
	return driver.Reconcile(ctx)
}
