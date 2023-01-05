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
	"fr.esgi/ha-audit/api/v1beta1"
	appsv1beta1 "fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	ha_service "fr.esgi/ha-audit/controllers/pkg/service"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	_ "sigs.k8s.io/controller-runtime/examples/crd/pkg"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		kernel.Logger.Error(err, "unable to fetch HA Audit CRD")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var service *ha_service.HAAuditService
	service = ha_service.New(r.Client, ctx, &haAudit)
	_, err := service.ApplyStrategy()
	if err != nil {
		return ctrl.Result{}, err
	}

	/*
		//https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html
		isUnderDeletion := !(haAudit.ObjectMeta.DeletionTimestamp.IsZero())
		thereIsFinalizer := controllerutil.ContainsFinalizer(&haAudit, finalizerName)
		if isUnderDeletion {
			if thereIsFinalizer {
				// Remove resources
				/*
					if err := mongoService.Delete(); err != nil {
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
	*/

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *HAAuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.HAAudit{}).
		Complete(r)
}
