package resources

import (
	"fr.esgi/ha-audit/api/v1beta1"
	v1api "k8s.io/api/core/v1"
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

func GetResourceTypePayload(resourceType v1beta1.AuditTargetType) ResourceTypePayload {
	switch resourceType {
	case v1beta1.PodTarget:
		return ResourceTypePayload{"pods", "v1", ""}
	case v1beta1.DeploymentTarget:
		return ResourceTypePayload{"deployments", "v1", "apps"}
	case v1beta1.StatefulSetTarget:
		return ResourceTypePayload{"statefulsets", "v1", "apps"}
	case v1beta1.DaemonSetTarget:
		return ResourceTypePayload{"daemonsets", "v1", "apps"}
	case v1beta1.ReplicaSetTarget:
		return ResourceTypePayload{"replicasets", "v1", "apps"}
	}
	return ResourceTypePayload{}
}
