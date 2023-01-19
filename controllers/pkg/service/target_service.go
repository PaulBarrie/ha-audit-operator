package service

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	resource_repo "fr.esgi/ha-audit/controllers/pkg/repository/resources"
	"net/http"
)

func (H *HAAuditService) _acquireTargets() ([]resource_repo.TargetResourcePayload, error) {
	get, err := H.ResourceRepository.GetAll(H.CRD.Spec.Targets)
	if err != nil {
		kernel.Logger.Error(err, "unable to get targets")
		return []resource_repo.TargetResourcePayload{}, err
	}
	return get, nil
}

func _testTarget(target v1beta1.Target) (bool, error) {
	response, err := http.Get(target.Path)
	if err != nil {
		kernel.Logger.Info(fmt.Sprintf("Unable to test target: %v", err))
		return false, err
	}
	if response.StatusCode/100 == 5 {
		return false, nil
	}
	return true, nil
}

func (H *HAAuditService) _inferTargets() []v1beta1.Target {
	var inferredTargets []v1beta1.Target
	for _, target := range H.CRD.Spec.Targets {
		if target.Id == "" || target.Path == "" || target.Name == "" {
			inferredTargets = append(inferredTargets, target.Default(H.CRD.ObjectMeta.Namespace))
		}
	}
	return inferredTargets
}
