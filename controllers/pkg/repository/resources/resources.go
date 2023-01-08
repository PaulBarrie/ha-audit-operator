package resources

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	v1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/applyconfigurations/core/v1"
)

type TargetResourcePayload struct {
	TargetType v1beta1.AuditTargetType
	Pods       []v1api.Pod
	Id         string
}
type ResourceTypePayload struct {
	Kind       string
	ApiVersion string
	Group      string
}

type ResourceAPIName string

const (
	PodResourceAPIName         ResourceAPIName = "pods"
	DeploymentResourceAPIName                  = "deployments"
	StatefulsetResourceAPIName                 = "statefulsets"
	DaemonsetResourceAPIName                   = "daemonsets"
	ReplicasetResourceAPIName                  = "replicasets"
)

func GetResourceTypePayload(resourceType v1beta1.AuditTargetType) schema.GroupVersionResource {
	switch resourceType {
	case v1beta1.PodTarget:
		return schema.GroupVersionResource{Version: "v1", Resource: string(PodResourceAPIName)}
	case v1beta1.DeploymentTarget:
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case v1beta1.StatefulsetTarget:
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	case v1beta1.DaemonsetTarget:
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	case v1beta1.ReplicasetTarget:
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}
	}
	kernel.Logger.Error(
		kernel.ErrorNotFound(
			fmt.Sprintf("%s is not a valid target type", resourceType)),
		"resourceType")
	return schema.GroupVersionResource{}
}
