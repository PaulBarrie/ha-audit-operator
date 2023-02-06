package service

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"fr.esgi/ha-audit/controllers/pkg/repository/cron_cache"
	"github.com/robfig/cron/v3"
	"reflect"
)

type PrometheusReportInitPayload struct {
	Report  v1beta1.PrometheusReport
	IdTotal string
	IdRate  string
}

func (H *HAAuditService) initPrometheusReport(crd *v1beta1.HAAudit) PrometheusReportInitPayload {
	totalInstanceUp := *v1beta1.DefaultTotalRunningInstanceMetric(
		crd.ObjectMeta.Name,
		int(crd.Spec.TestScheduleSeconds),
	)
	idTot, err1 := H.PrometheusRepository.Create(totalInstanceUp)
	if err1 != nil {
		kernel.Logger.Error(err1, "unable to update prometheus")
	}
	H.EventRecorder.Event(
		crd,
		totalInstanceUp.MetricEventPayload.Type,
		totalInstanceUp.MetricEventPayload.Reason,
		totalInstanceUp.MetricEventPayload.Message,
	)

	rateInstanceUp := *v1beta1.DefaultTotalRunningInstanceRateMetric(
		crd.ObjectMeta.Name,
		int(crd.Spec.TestScheduleSeconds),
	)
	idRate, err2 := H.PrometheusRepository.Create(rateInstanceUp)
	if err2 != nil {
		kernel.Logger.Error(err2, "unable to update prometheus")
	}
	H.EventRecorder.Event(
		crd,
		rateInstanceUp.MetricEventPayload.Type,
		rateInstanceUp.MetricEventPayload.Reason,
		rateInstanceUp.MetricEventPayload.Message,
	)

	return PrometheusReportInitPayload{
		Report: v1beta1.PrometheusReport{
			DumpFrequencySeconds: crd.Spec.Report.PrometheusReport.DumpFrequencySeconds,
		},
		IdTotal: idTot.(string),
		IdRate:  idRate.(string),
	}
}

func (H *HAAuditService) _getTestFunctionCron(crd *v1beta1.HAAudit) func() {
	return func() {
		nbServiceUp := 0
		for _, target := range crd.Spec.Targets {
			ok, err := _testTarget(target)
			if err != nil {
				kernel.Logger.Info(fmt.Sprintf("Unable to test target %s", target.Name))
			} else if ok {
				nbServiceUp++
			}
		}
		//if reflect.DeepEqual(crd.Spec.Report.PrometheusReport, v1beta1.PrometheusReport{}) {
		//	H.initPrometheusReport(crd)
		//}
		if err := H.PrometheusRepository.Update(
			crd.Status.MetricStatus.TotalUpMetricID, float64(nbServiceUp)); err != nil {
			kernel.Logger.Error(err, "unable to update prometheus")
		}
		if err := H.PrometheusRepository.Update(
			crd.Status.MetricStatus.RateUpMetricID,
			float64(nbServiceUp/len(crd.Spec.Targets))); err != nil {
			kernel.Logger.Error(err, "unable to update prometheus")
		}
	}
}

func (H *HAAuditService) _scheduleTestReport(crd *v1beta1.HAAudit) (error, cron.EntryID) {
	var newCRD = crd.DeepCopy()

	doesNotExist := !(crd.Status.Created && cron_cache.GetDB().Exists(crd.Status.TestStatus.CronID))
	shouldBeUpdate := crd.Spec.TestScheduleSeconds != newCRD.Spec.TestScheduleSeconds || !reflect.DeepEqual(crd.Spec.Targets, newCRD.Spec.Targets)

	if doesNotExist {
		kernel.Logger.Info("Create test report cron")
		cronId, err := H.CronRepository.Create(int(crd.Spec.TestScheduleSeconds), H._getTestFunctionCron(crd))
		if err != nil {
			return err, cron.EntryID(0)
		}
		return nil, cronId.(cron.EntryID)
	} else if shouldBeUpdate {
		cronId, err := H.CronRepository.Update(crd.Status.TestStatus.CronID, int(crd.Spec.TestScheduleSeconds), H._getTestFunctionCron(crd))
		if err != nil {
			return err, cron.EntryID(0)
		}
		crd.Status.TestStatus.CronID = cronId.(cron.EntryID)
	}
	return nil, cron.EntryID(0)
}
