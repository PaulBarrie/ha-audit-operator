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

func (H *HAAuditService) _getStrategyCronFunction(crd *v1beta1.HAAudit) (func(), error) {
	var cronFunction func()

	delTarget := func(target resource_repo.TargetResourcePayload) {
		kernel.Logger.Info(fmt.Sprintf("deleting target %s", target.TargetType))
		for i := 0; i < crd.Spec.ChaosStrategy.NumberOfPodsToKill; i++ {
			podToDeleteIndex := rand.IntnRange(0, len(target.Pods)-1)
			pod := target.Pods[podToDeleteIndex]
			if err := H.ResourceRepository.Delete(&pod); err != nil {
				kernel.Logger.Error(err, "unable to delete pod")
			}
		}
		crd.Status.NextChaosDateTime = time.Now().Unix() + crd.Spec.ChaosStrategy.FrequencySeconds
	}

	switch crd.Spec.ChaosStrategy.ChaosStrategyType {
	case v1beta1.ChaosStrategyTypeRandom:
		cronFunction = func() {
			targets, err := H._acquireTargets(crd)
			if err != nil {
				kernel.Logger.Error(err, "unable to acquire targets")
				return
			}
			targetsFlat := resource_repo.TargetResourcePayload{TargetType: v1beta1.PodTarget, Pods: []v1api.Pod{}}
			for _, target := range targets {
				targetsFlat.Pods = append(targetsFlat.Pods, target.Pods...)
			}
			if err != nil {
				kernel.Logger.Error(err, "unable to acquire targets")
				return
			}
			delTarget(targetsFlat)
		}
	case v1beta1.ChaosStrategyTypeRoundRobin:
		cronFunction = func() {
			targets, err := H._acquireTargets(crd)
			if err != nil {
				kernel.Logger.Error(err, "unable to acquire targets")
				return
			}
			var targetCandidate resource_repo.TargetResourcePayload
			var targetIndex int
			for i, target := range targets {
				targetIndex = i
				if target.Id == crd.Status.RoundRobinStrategy.CurrentTargetId {
					targetCandidate = target
					break
				}
			}
			crd.Status.RoundRobinStrategy.CurrentTargetId = targets[(targetIndex+1)%len(targets)].Id
			delTarget(targetCandidate)
		}
	case v1beta1.ChaosStrategyTypeFixed:
		cronFunction = func() {
			targets, err := H._acquireTargets(crd)
			if err != nil {
				kernel.Logger.Error(err, "unable to acquire targets")
				return
			}
			for _, target := range targets {
				delTarget(target)
			}
		}
	default:
		return func() {}, kernel.ErrorDoesNotExist(fmt.Sprintf("Strategy %s does not exist", crd.Spec.ChaosStrategy.ChaosStrategyType))
	}

	return cronFunction, nil
}

func (H *HAAuditService) _scheduleStrategy(crd *v1beta1.HAAudit) (error, cron.EntryID) {
	kernel.Logger.Info("scheduling strategy")
	targets, err := H._acquireTargets(crd)
	if err != nil || len(targets) == 0 {
		kernel.Logger.Error(err, "unable to get targets")
		return err, cron.EntryID(0)
	}
	cronFunc, errStrat := H._getStrategyCronFunction(crd)
	if errStrat != nil {
		kernel.Logger.Error(errStrat, "unable to get cron function")
		return errStrat, cron.EntryID(0)
	}
	cronId, errCron := H.CronRepository.Create(int(crd.Spec.ChaosStrategy.FrequencySeconds), cronFunc)
	if errCron != nil {
		kernel.Logger.Error(errCron, "unable to create cron")
		return errCron, cron.EntryID(0)
	}
	idCron := cronId.(cron.EntryID)
	return nil, idCron
}
