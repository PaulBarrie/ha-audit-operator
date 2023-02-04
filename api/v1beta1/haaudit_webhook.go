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

package v1beta1

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var haauditlog = logf.Log.WithName("haaudit-resource")

func (r *HAAudit) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-apps-fr-esgi-v1beta1-haaudit,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.fr.esgi,resources=haaudits,verbs=create;update,versions=v1beta1,name=mhaaudit.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &HAAudit{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HAAudit) Default() {
	haauditlog.Info("default", "name", r.Name)
	newTargets := make([]Target, 0)
	for _, target := range r.Spec.Targets {
		newTargets = append(newTargets, target.Default(r.Namespace))
	}
	r.Spec.Targets = newTargets
	r.Spec.ChaosStrategy.Default()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-fr-esgi-v1beta1-haaudit,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.fr.esgi,resources=haaudits,verbs=create;update,versions=v1beta1,name=vhaaudit.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &HAAudit{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HAAudit) ValidateCreate() error {
	haauditlog.Info("validate create", "name", r.Name)
	return r._validateCRD()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HAAudit) ValidateUpdate(old runtime.Object) error {
	haauditlog.Info("validate update", "name", r.Name)

	return r._validateCRD()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HAAudit) ValidateDelete() error {
	haauditlog.Info("validate delete", "name", r.Name)
	return r._validateCRD()
}

func (r *HAAudit) _validateCRD() error {
	if r.Spec.Targets == nil || len(r.Spec.Targets) == 0 {
		return fmt.Errorf("targets cannot be nil")
	}
	for _, target := range r.Spec.Targets {
		if target.Name == "" && target.Namespace == "" && target.NameRegex == "" {
			return fmt.Errorf("target name cannot be empty")
		}
		if target.Namespace == "" {
			return fmt.Errorf("target namespace cannot be empty")
		}
	}
	return nil
}
