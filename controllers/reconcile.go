package controllers

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	clog "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	log = clog.Log.WithName("generic-reconciler")
)

// Params is a parameter object for the ReconcileResources function
type Params struct {
	Client client.Client
	// Owner will be set as the controller reference
	Owner client.Object
	// Expected the expected state of the resource going into reconciliation.
	Expected client.Object
	// Reconciled will contain the final state of the resource after reconciliation containing the
	// unification of remote and expected state.
	Reconciled client.Object
	// NeedsUpdate returns true when the object to be reconciled has changes that are not persisted remotely.
	//NeedsUpdate func() bool
	// NeedsRecreate returns true when the object to be reconciled needs to be deleted and re-created because it cannot be updated.
	//NeedsRecreate func() bool
	// UpdateReconciled modifies the resource pointed to by Reconciled to reflect the state of Expected
	UpdateReconciled func()
	// PreCreate is called just before the creation of the resource.
	PreCreate func() error
	// PreUpdate is called just before the update of the resource.
	PreUpdate func() error
	// PostUpdate is called immediately after the resource is successfully updated.
	//PostUpdate func()
}

func (p Params) CheckNilValues() error {
	if p.Reconciled == nil {
		return errors.New("Reconciled must not be nil")
	}
	//if p.UpdateReconciled == nil {
	//	return errors.New("UpdateReconciled must not be nil")
	//}
	//if p.NeedsUpdate == nil {
	//	return errors.New("NeedsUpdate must not be nil")
	//}
	if p.Expected == nil {
		return errors.New("Expected must not be nil")
	}
	return nil
}

// ReconcileResource is a generic reconciliation function for resources that need to
// implement runtime.Object and meta/v1.Object.
func ReconcileResource(params Params) error {
	err := params.CheckNilValues()
	if err != nil {
		return err
	}
	namespace := params.Expected.GetNamespace()
	name := params.Expected.GetName()
	gvk, err := apiutil.GVKForObject(params.Expected, scheme.Scheme)
	if err != nil {
		return err
	}
	kind := gvk.Kind

	//if params.Owner != nil {
	//	if err := controllerutil.SetControllerReference(params.Owner, params.Expected, scheme.Scheme); err != nil {
	//		return err
	//	}
	//}

	create := func() error {
		log.Info("Creating resource", "kind", kind, "namespace", namespace, "name", name)
		if params.PreCreate != nil {
			if err := params.PreCreate(); err != nil {
				return err
			}
		}

		expectedCopyValue := reflect.ValueOf(params.Expected.DeepCopyObject()).Elem()
		reflect.ValueOf(params.Reconciled).Elem().Set(expectedCopyValue)
		// Create the object, which modifies params.Reconciled in-place
		err = params.Client.Create(context.Background(), params.Reconciled)
		if err != nil {
			return err
		}
		return nil
	}

	// Check if already exists,Create if not present
	err = params.Client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, params.Reconciled)
	if err != nil && apierrors.IsNotFound(err) {
		return create()
	} else if err != nil {
		log.Error(err, fmt.Sprintf("Generic GET for %s %s/%s failed with error", kind, namespace, name))
		return fmt.Errorf("failed to get %s %s/%s: %w", kind, namespace, name, err)
	}

	return nil
}
