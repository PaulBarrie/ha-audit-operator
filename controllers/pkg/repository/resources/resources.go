package resources

import (
	"fr.esgi/ha-audit/api/v1beta1"
	_ "k8s.io/client-go/applyconfigurations/core/v1"
)

type ResourceTypePayload struct {
	Kind       string
	ApiVersion string
	Group      string
}

func GetResourceTypePayload(resourceType v1beta1.AuditTargetType) ResourceTypePayload {
	switch resourceType {
	case v1beta1.Pod:
		return ResourceTypePayload{"pods", "v1", ""}
	case v1beta1.Deployment:
		return ResourceTypePayload{"deployments", "v1", "apps"}
	case v1beta1.StatefulSet:
		return ResourceTypePayload{"statefulsets", "v1", "apps"}
	case v1beta1.DaemonSet:
		return ResourceTypePayload{"daemonsets", "v1", "apps"}
	case v1beta1.ReplicaSet:
		return ResourceTypePayload{"replicasets", "v1", "apps"}
	}
	return ResourceTypePayload{}
}
