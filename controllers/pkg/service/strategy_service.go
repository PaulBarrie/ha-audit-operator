package service

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	resource_repo "fr.esgi/ha-audit/controllers/pkg/repository/resources"
	"github.com/robfig/cron/v3"
	v1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

func (H *HAAuditService) _getStrategyCronFunction(targets []resource_repo.TargetResourcePayload) (func(), error) {
	var cronFunction func()

	delTarget := func(target resource_repo.TargetResourcePayload) {
		if H.CRD.Status.NextChaosDateTime < time.Now().Unix() {
			return
		}
		kernel.Logger.Info(fmt.Sprintf("deleting target %s", target.TargetType))
		for i := 0; i < H.CRD.Spec.ChaosStrategy.NumberOfPodsToKill; i++ {
			podToDeleteIndex := rand.IntnRange(0, len(targets))
			pod := target.Pods[podToDeleteIndex]
			if err := H.ResourceRepository.Delete(&pod); err != nil {
				kernel.Logger.Error(err, "unable to delete pod")
			}
			targets = append(targets[:podToDeleteIndex], targets[podToDeleteIndex+1:]...)
		}
		H.CRD.Status.NextChaosDateTime = time.Now().Unix() + H.CRD.Spec.ChaosStrategy.FrequencySeconds
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
				if target.Id == H.CRD.Status.RoundRobinStrategy.CurrentTargetId {
					targetCandidate = target
					break
				}
			}
			newCRD := H.CRD.DeepCopy()
			newCRD.Status.RoundRobinStrategy.CurrentTargetId = targets[(targetIndex+1)%len(targets)].Id
			err := H.CRDRepository.Update(newCRD)
			if err != nil {
				return
			}
			delTarget(targetCandidate)
		}
	case v1beta1.ChaosStrategyTypeFixed:
		cronFunction = func() {
			for _, target := range targets {
				delTarget(target)
			}
		}
	default:
		return func() {}, kernel.ErrorDoesNotExist(fmt.Sprintf("Strategy %s does not exist", H.CRD.Spec.ChaosStrategy.ChaosStrategyType))
	}

	return cronFunction, nil
}

func (H *HAAuditService) _scheduleStrategy() error {
	kernel.Logger.Info("scheduling strategy")
	targets, err := H._acquireTargets()
	if err != nil || len(targets) == 0 {
		kernel.Logger.Error(err, "unable to get targets")
		return err
	}
	cronFunc, err := H._getStrategyCronFunction(targets)
	if err != nil {
		kernel.Logger.Error(err, "unable to get cron function")
		return err
	}
	cronId, err := H.CronRepository.Create(int(H.CRD.Spec.ChaosStrategy.FrequencySeconds), cronFunc)
	if err != nil {
		kernel.Logger.Error(err, "unable to create cron")
		return err
	}
	H.CRD.Spec.ChaosStrategy.CronId = int(cronId.(cron.EntryID))
	return nil
}
