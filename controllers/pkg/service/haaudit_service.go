package service

import (
	"context"
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	cron_repo "fr.esgi/ha-audit/controllers/pkg/repository/cron"
	"fr.esgi/ha-audit/controllers/pkg/repository/prometheus"
	resource_repo "fr.esgi/ha-audit/controllers/pkg/repository/resources"
	v1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodTargets struct {
	Targets []v1beta1.Target
	Objects []v1api.Pod
}

type HAAuditService struct {
	CRD                  *v1beta1.HAAudit
	Client               client.Client
	Context              context.Context
	Targets              []resource_repo.TargetResourcePayload
	ResourceRepository   *resource_repo.ResourceRepository
	CronRepository       *cron_repo.CronRepository
	PrometheusRepository *prometheus.PrometheusRepository
}

func New(client client.Client, ctx context.Context, crd *v1beta1.HAAudit) *HAAuditService {
	return &HAAuditService{
		CRD:                crd,
		Client:             client,
		Context:            ctx,
		Targets:            []resource_repo.TargetResourcePayload{},
		ResourceRepository: resource_repo.GetInstance(client, ctx),
		CronRepository:     cron_repo.GetInstance(),
	}
}

func (H *HAAuditService) _acquireTargets() ([]resource_repo.TargetResourcePayload, error) {
	get, err := H.ResourceRepository.GetAll(H.CRD.Spec.Targets)
	if err != nil {
		kernel.Logger.Error(err, "unable to get targets")
		return []resource_repo.TargetResourcePayload{}, err
	}
	return get, nil
}

func (H *HAAuditService) ApplyStrategy() (int, error) {
	var cronFunction func()
	targets, err := H._acquireTargets()
	if err != nil {
		kernel.Logger.Error(err, "unable to get targets")
		return -1, err
	}
	delTarget := func(target resource_repo.TargetResourcePayload) {
		for i := 0; i < H.CRD.Spec.ChaosStrategy.NumberOfPodsToKill; i++ {
			podToDeleteIndex := rand.IntnRange(0, len(targets))
			pod := target.Pods[podToDeleteIndex]
			if err := H.ResourceRepository.Delete(&pod); err != nil {
				kernel.Logger.Error(err, "unable to delete pod")
			}
			targets = append(targets[:podToDeleteIndex], targets[podToDeleteIndex+1:]...)
		}
	}

	switch H.CRD.Spec.ChaosStrategy.ChaosStrategyType {
	case v1beta1.ChaosStrategyTypeRandom:
		targetsFlat := resource_repo.TargetResourcePayload{TargetType: v1beta1.PodTarget, Pods: []v1api.Pod{}}
		for _, target := range targets {
			targetsFlat.Pods = append(targetsFlat.Pods, target.Pods...)
		}
		cronFunction = func() {
			delTarget(targetsFlat)
		}
	case v1beta1.ChaosStrategyTypeRoundRobin:
		cronFunction = func() {
			var targetCandidate resource_repo.TargetResourcePayload
			var targetIndex int
			for i, target := range targets {
				targetIndex = i
				if target.Id == H.CRD.Spec.ChaosStrategy.RoundRobinStrategy.CurrentTargetId {
					targetCandidate = target
					break
				}
			}
			H.CRD.Spec.ChaosStrategy.RoundRobinStrategy.CurrentTargetId = targets[(targetIndex+1)%len(targets)].Id
			// Save CRD
			delTarget(targetCandidate)
		}
	case v1beta1.ChaosStrategyTypeFixed:
		cronFunction = func() {
			for _, target := range targets {
				delTarget(target)
			}
		}
	default:
		return -1, nil
	}

	cronIds, err := H.CronRepository.Create(H.CRD.Spec.ChaosStrategy.FrequencyCron, cronFunction)
	if err != nil {
		kernel.Logger.Error(err, "unable to create cron")
		return -1, err
	}
	return cronIds.(int), nil
}

func (H *HAAuditService) _scheduleTests() (int, error) {
	cronId, err := H.CronRepository.Create(H.CRD.Spec.TestSchedule, func() {
		nbServiceUp := 0
		for _, target := range H.CRD.Spec.Targets {
			ok, err := _testTarget(target)
			if err != nil {
				kernel.Logger.Error(err, "unable to test target")
			} else if ok {
				nbServiceUp++
			}
		}
		if !reflect.DeepEqual(H.CRD.Spec.Report.PrometheusReport, v1beta1.PrometheusReport{}) {
			err := H.PrometheusRepository.Update(H.CRD.Spec.Report.PrometheusReport.InstanceUp, nbServiceUp)
			if err != nil {
				kernel.Logger.Error(err, "unable to update prometheus")
			}
		}
		if !reflect.DeepEqual(H.CRD.Spec.Report.GrafanaReport, v1beta1.GrafanaReport{}) {
			//err = H.PrometheusRepository.Update(H.CRD.Spec.HAReport.GrafanaReport., nbServiceUp)
			//if err != nil {
			//	kernel.Logger.Error(err, "unable to update prometheus")
			//}
		}

	})
	if err != nil {
		kernel.Logger.Error(err, "unable to create cron")
		return -1, err
	}
	return cronId.(int), nil
}

func _testTarget(target v1beta1.Target) (bool, error) {
	response, err := http.Get(target.Path)
	if response.StatusCode/100 == 5 {
		return false, nil
	} else if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return false, err
	}
	return true, nil
}
