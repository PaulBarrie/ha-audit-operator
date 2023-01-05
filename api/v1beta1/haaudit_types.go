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
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AuditTargetType string

var (
	PodTarget         AuditTargetType = "pod"
	DeploymentTarget  AuditTargetType = "deployment"
	StatefulSetTarget AuditTargetType = "statefulset"
	DaemonSetTarget   AuditTargetType = "daemonset"
	ReplicaSetTarget  AuditTargetType = "replicaset"
)

type Target struct {
	Id   string          `json:"id"`
	Kind AuditTargetType `json:"kind"`
	// +optional
	Name string `json:"name"`
	// +optional
	NameRegex string `json:"nameRegex"`
	// +optional
	Namespace string `json:"namespace"`
	// +optional
	LabelSelector metav1.LabelSelector `json:"labelSelector"`
	// +optional
	// +kubebuilder:default:="/"
	Path string `json:"path"`
}

func (t *Target) Default(namespace string) Target {
	if t.Namespace == "" {
		t.Namespace = namespace
	}
	if t.Path == "" {
		t.Path = "/"
	}
	if t.Id == "" {
		id, err := uuid.NewRandom()
		if err != nil {
			kernel.Logger.Error(err, "Error while generating UUID")
		}
		t.Id = id.String()
	}
	return *t
}

// HAAuditSpec defines the desired state of HAAudit
type HAAuditSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Targets []Target `json:"targets"`
	// +kubebuilder:validation:Required
	ChaosStrategy ChaosStrategy `json:"chaosStrategy"`

	// +optional
	// +kubebuilder:default:="* * * * *"
	TestSchedule string `json:"testSchedule"`
	// +optional
	Report HAReport `json:"report"`
	// +kubebuilder:validation:ExclusiveMinimum=0
	StrategyCronList []int `json:"cronList"`
	// +kubebuilder:validation:ExclusiveMinimum=0
	ReportCronList []int `json:"cronList"`
}

// HAAuditStatus defines the observed state of HAAudit
type HAAuditStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HAAudit is the Schema for the haaudits API
type HAAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HAAuditSpec   `json:"spec,omitempty"`
	Status HAAuditStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HAAuditList contains a list of HAAudit
type HAAuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HAAudit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HAAudit{}, &HAAuditList{})
}
