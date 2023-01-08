package service

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"github.com/robfig/cron/v3"
	"reflect"
)

func (H *HAAuditService) initPrometheusReport() {
	if H.PrometheusRepository.Address == "" {
		kernel.Logger.Info("PrometheusRepository is nil")
		return
	}
	H.CRD.Spec.Report.PrometheusReport.InstanceUp = *v1beta1.DefaultTotalRunningInstanceMetric(
		H.CRD.ObjectMeta.Name,
		int(H.CRD.Spec.TestScheduleSeconds),
	)
	if _, err := H.PrometheusRepository.Create(
		H.CRD.Spec.Report.PrometheusReport.InstanceUp,
	); err != nil {
		kernel.Logger.Error(err, "unable to update prometheus")
	}
	H.CRD.Spec.Report.PrometheusReport.InstanceUpRate = *v1beta1.DefaultTotalRunningInstanceRateMetric(
		H.CRD.ObjectMeta.Name,
		int(H.CRD.Spec.TestScheduleSeconds),
	)
	if _, err := H.PrometheusRepository.Create(
		H.CRD.Spec.Report.PrometheusReport.InstanceUpRate, 0); err != nil {
		kernel.Logger.Error(err, "unable to update prometheus")
	}
}

func (H *HAAuditService) _getTestFunctionCron() func() {
	return func() {
		nbServiceUp := 0
		for _, target := range H.CRD.Spec.Targets {
			ok, err := _testTarget(target)
			if err != nil {
				kernel.Logger.Info(fmt.Sprintf("Unable to test target %s", target.Name))
			} else if ok {
				nbServiceUp++
			}
		}
		if H.PrometheusRepository.Address == "" {
			kernel.Logger.Info("PrometheusRepository is nil")
			return
		}
		H.CRD.Spec.Report.PrometheusReport = H.CRD.Spec.Report.PrometheusReport.Get(
			H.CRD.ObjectMeta.Name,
			H.CRD.Spec.Report.PrometheusReport.DumpFrequencySeconds,
		)
		if _, err := H.PrometheusRepository.Update(
			H.CRD.Spec.Report.PrometheusReport.InstanceUp, float64(nbServiceUp)); err != nil {
			kernel.Logger.Error(err, "unable to update prometheus")
		}
		if _, err := H.PrometheusRepository.Update(
			H.CRD.Spec.Report.PrometheusReport.InstanceUpRate,
			float64(nbServiceUp/len(H.CRD.Spec.Targets))); err != nil {
			kernel.Logger.Error(err, "unable to update prometheus")
		}
		//if err := H.Client.Update(H.Context, H.CRD); err != nil {
		//	kernel.Logger.Error(err, "unable to update crd")
		//}
	}
}

func (H *HAAuditService) _scheduleTestReport() error {
	var newCRD = H.CRD.DeepCopy()

	doesNotExist := H.CRD.Spec.TestReportCronId == 0
	shouldBeUpdate := H.CRD.Spec.TestScheduleSeconds != newCRD.Spec.TestScheduleSeconds || !reflect.DeepEqual(H.CRD.Spec.Targets, newCRD.Spec.Targets)
	if doesNotExist {
		kernel.Logger.Info("Create test report cron")
		cronId, err := H.CronRepository.Create(int(H.CRD.Spec.TestScheduleSeconds), H._getTestFunctionCron())
		if err != nil {
			return err
		}
		H.CRD.Spec.TestReportCronId = cronId.(cron.EntryID)
	} else if shouldBeUpdate {
		cronId, err := H.CronRepository.Update(H.CRD.Spec.TestReportCronId, int(H.CRD.Spec.TestScheduleSeconds), H._getTestFunctionCron())
		if err != nil {
			return err
		}
		H.CRD.Spec.TestReportCronId = cronId.(cron.EntryID)
	}
	return nil
}
