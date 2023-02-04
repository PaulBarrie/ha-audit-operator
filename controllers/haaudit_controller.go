/*
Copyright 2022.

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
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	appsv1beta1 "fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	ha_service "fr.esgi/ha-audit/controllers/pkg/service"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	_ "sigs.k8s.io/controller-runtime/examples/crd/pkg"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	finalizerName = "haaudit.finalizers.esgi.fr"
)

// HAAuditReconciler reconciles a HAAudit object
type HAAuditReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.fr.esgi,resources=haaudits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.fr.esgi,resources=haaudits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.fr.esgi,resources=haaudits/finalizers,verbs=update

//+kubebuilder:rbac:groups=*,resources=deployments;pods;statefulsets;replicasets;daemonsets,verbs=get;list;create;update;watch;delete

//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HAAudit object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *HAAuditReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	kernel.Logger.WithValues("Namespace", req.NamespacedName)
	haAudit := v1beta1.HAAudit{}
	if err := r.Get(ctx, req.NamespacedName, &haAudit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var service *ha_service.HAAuditService
	service = ha_service.New(r.Client, ctx, &haAudit)

	//https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html
	isUnderDeletion := !(haAudit.ObjectMeta.DeletionTimestamp.IsZero())
	thereIsFinalizer := controllerutil.ContainsFinalizer(&haAudit, finalizerName)

	if isUnderDeletion {
		kernel.Logger.Info("Deleting HA Audit CRD")
		if thereIsFinalizer {
			// Remove resources
			if err := service.Delete(); err != nil {
				return ctrl.Result{}, err
			}
			// Remove finalizer
			controllerutil.RemoveFinalizer(&haAudit, finalizerName)
			if err := r.Update(ctx, &haAudit); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if !thereIsFinalizer {
			controllerutil.AddFinalizer(&haAudit, finalizerName)
			if err := r.Update(ctx, &haAudit); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	kernel.Logger.Info("Reconciling HA Audit CRD")
	err, upCRD := service.CreateOrUpdate()
	kernel.Logger.Info(fmt.Sprintf("TOUP_HA: %v", upCRD.Status))
	if err != nil {
		return ctrl.Result{}, err
	}
	if err = r.Client.Update(ctx, upCRD); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err = r.Status().Update(ctx, upCRD); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	kernel.Logger.Info(fmt.Sprintf("HA Audit CRD reconciled : %v", upCRD.Status))
	Payload := `{ "op": "replace", "path": "/status/created","value": "true"}`
	bytePayload := []byte(Payload)
	err = r.Patch(ctx, upCRD, client.RawPatch(types.MergePatchType, bytePayload))
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HAAuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.HAAudit{}).
		Complete(r)
}
