package kernel

import "k8s.io/apimachinery/pkg/types"

func AddIfNotExists(targets []types.NamespacedName, new types.NamespacedName) []types.NamespacedName {
	for _, target := range targets {
		if target.Name == new.Name && target.Namespace == new.Namespace {
			return targets
		}
	}
	return append(targets, new)
}
