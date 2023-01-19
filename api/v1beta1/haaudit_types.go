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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	DefaultPathPrefix = "http://"
)

type AuditTargetType string

var (
	PodTarget         AuditTargetType = "pod"
	DeploymentTarget  AuditTargetType = "deployment"
	StatefulsetTarget AuditTargetType = "statefulset"
	DaemonsetTarget   AuditTargetType = "daemonset"
	ReplicasetTarget  AuditTargetType = "replicaset"
)

type Target struct {
	// +kubebuilder:validation:Optional
	Id string `json:"id"`
	// +kubebuilder:validation=Required
	Kind AuditTargetType `json:"kind"`
	// +kubebuilder:validation:Optional
	Name string `json:"name"`
	// +kubebuilder:validation:Optional
	NameRegex string `json:"nameRegex"`
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Optional
	LabelSelector metav1.LabelSelector `json:"labelSelector"`
	// +kubebuilder:validation:Required
	Path string `json:"path"`
}

func (t *Target) Default(namespace string) Target {
	if t.Namespace == "" {
		t.Namespace = namespace
	}
	if t.Id == "" {
		id, err := uuid.NewRandom()
		if err != nil {
			kernel.Logger.Error(err, "Error while generating UUID")
		}
		t.Id = id.String()
	}
	if !strings.HasPrefix(t.Path, DefaultPathPrefix) {
		t.Path = DefaultPathPrefix + t.Path
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

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=10
	TestScheduleSeconds int64 `json:"testSchedule"`

	// +kubebuilder:validation:Required
	Report HAReport `json:"report"`
}

type StrategyStatus struct {
	Cron   cron.EntryID     `json:"cron"`
	Metric prometheus.Gauge `json:"metric"`
}

type TestStatus struct {
	CronID cron.EntryID     `json:"cron"`
	Metric prometheus.Gauge `json:"metric"`
}

// HAAuditStatus defines the observed state of HAAudit
type HAAuditStatus struct {
	ChaosStrategyCron            cron.EntryID       `json:"chaosStrategyCron"`
	TestReportCron               cron.EntryID       `json:"testReportCron"`
	RoundRobinStrategy           RoundRobinStrategy `json:"roundRobinStrategy,omitempty"`
	FixedStrategy                FixedStrategy      `json:"fixedStrategy,omitempty"`
	PrometheusClusterRoleBinding NamespacedName     `json:"prometheusClusterRoleBinding"`
	NextChaosDateTime            int64              `json:"nextChaosDateTime"`
	Created                      bool               `json:"created,default=false"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HAAudit is the Schema for the haaudits API
type HAAudit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HAAuditSpec   `json:"spec,omitempty"`
	Status            HAAuditStatus `json:"status,omitempty"`
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
