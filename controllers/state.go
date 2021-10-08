package controllers

import (
	"context"
	"github.com/dataworkbench/hdfs-operator/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// State holds the accumulated state during the reconcile loop including the response and a pointer to an HDFS resource for status updates.
type State struct {
	//*events.Recorder
	cluster v1.HDFS
	status  v1.HDFSStatus
}

// NewState creates a new reconcile state based on the given cluster
func NewState(c v1.HDFS) *State {
	return &State{cluster: c, status: *c.Status.DeepCopy()}
}

// Results collects intermediate results of a reconciliation run and any errors that occurred.
type Results struct {
	currResult reconcile.Result //controller-runtime  type Result struct { Requeue  RequeueAfter }
	currKind   resultKind
	errors     []error
	ctx        context.Context
}

type resultKind int

// WithError adds an error to the results.
func (r *Results) WithError(err error) *Results {
	if err != nil {
		//r.errors = append(r.errors, tracing.CaptureError(r.ctx, err))
		r.errors = append(r.errors, err)
	}
	return r
}

// Aggregate returns the highest priority reconcile result and any errors seen so far.
func (r *Results) Aggregate() (reconcile.Result, error) {
	return r.currResult, k8serrors.NewAggregate(r.errors)
}
