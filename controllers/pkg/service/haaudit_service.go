package service

import (
	"context"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"fr.esgi/ha-audit/controllers/pkg/repository/resources"
	v1api "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodTargets struct {
	Targets []v1beta1.Target
	Objects []v1api.Pod
}

type HAAuditService struct {
	CRD                *v1beta1.HAAudit
	Client             client.Client
	Context            context.Context
	Targets            PodTargets
	ResourceRepository resources.ResourceRepository
}

func New(client client.Client, ctx context.Context, crd *v1beta1.HAAudit) *HAAuditService {
	return &HAAuditService{
		CRD:                crd,
		Client:             client,
		Context:            ctx,
		Targets:            PodTargets{Targets: crd.Spec.Targets},
		ResourceRepository: resources.ResourceRepository{Context: ctx, Client: client},
	}
}

func (H *HAAuditService) GetTargets() (PodTargets, error) {
	for _, target := range H.CRD.Spec.Targets {
		get, err := H.ResourceRepository.Get(target)
		if err != nil {
			kernel.Logger.Error(err, "unable to fetch target")
			break
		}
		kernel.Logger.Info("target fetched", "target", target)
		for _, pod := range get {
			kernel.Logger.Info("pod found", "pod", pod.Name)
		}
		H.Targets.Objects = append(H.Targets.Objects, get...)
	}

	return H.Targets, nil
}
