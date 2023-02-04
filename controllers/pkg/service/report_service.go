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

func (H *HAAuditService) initPrometheusReport() PrometheusReportInitPayload {
	totalInstanceUp := *v1beta1.DefaultTotalRunningInstanceMetric(
		H.CRD.ObjectMeta.Name,
		int(H.CRD.Spec.TestScheduleSeconds),
	)
	idTot, err1 := H.PrometheusRepository.Create(totalInstanceUp)
	if err1 != nil {
		kernel.Logger.Error(err1, "unable to update prometheus")
	}

	rateInstanceUp := *v1beta1.DefaultTotalRunningInstanceRateMetric(
		H.CRD.ObjectMeta.Name,
		int(H.CRD.Spec.TestScheduleSeconds),
	)
	idRate, err2 := H.PrometheusRepository.Create(rateInstanceUp)
	if err2 != nil {
		kernel.Logger.Error(err2, "unable to update prometheus")
	}

	return PrometheusReportInitPayload{
		Report: v1beta1.PrometheusReport{
			DumpFrequencySeconds: H.CRD.Spec.Report.PrometheusReport.DumpFrequencySeconds,
			InstanceUp:           totalInstanceUp,
			InstanceUpRate:       rateInstanceUp,
		},
		IdTotal: idTot.(string),
		IdRate:  idRate.(string),
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

		if reflect.DeepEqual(H.CRD.Spec.Report.PrometheusReport, v1beta1.PrometheusReport{}) {
			H.initPrometheusReport()
		}
		//H.CRD.Spec.Report.PrometheusReport = H.CRD.Spec.Report.PrometheusReport.Get(
		//	H.CRD.ObjectMeta.Name,
		//	H.CRD.Spec.Report.PrometheusReport.DumpFrequencySeconds,
		//)
		kernel.Logger.Info(fmt.Sprintf("Metrics Status : %v", H.CRD.Status))
		kernel.Logger.Info(fmt.Sprintf("Update nb of svc up: %d", nbServiceUp))
		if err := H.PrometheusRepository.Update(
			H.CRD.Status.TestStatus.TotalUpMetricID, float64(nbServiceUp)); err != nil {
			kernel.Logger.Error(err, "unable to update prometheus")
		}
		if err := H.PrometheusRepository.Update(
			H.CRD.Status.TestStatus.RateUpMetricID,
			float64(nbServiceUp/len(H.CRD.Spec.Targets))); err != nil {
			kernel.Logger.Error(err, "unable to update prometheus")
		}
	}
}

func (H *HAAuditService) _scheduleTestReport() (error, cron.EntryID) {
	var newCRD = H.CRD.DeepCopy()

	doesNotExist := !(H.CRD.Status.Created && cron_cache.GetDB().Exists(H.CRD.Status.TestStatus.CronID))
	shouldBeUpdate := H.CRD.Spec.TestScheduleSeconds != newCRD.Spec.TestScheduleSeconds || !reflect.DeepEqual(H.CRD.Spec.Targets, newCRD.Spec.Targets)

	if doesNotExist {
		kernel.Logger.Info("Create test report cron")
		cronId, err := H.CronRepository.Create(int(H.CRD.Spec.TestScheduleSeconds), H._getTestFunctionCron())
		if err != nil {
			return err, cron.EntryID(0)
		}
		return nil, cronId.(cron.EntryID)
	} else if shouldBeUpdate {
		cronId, err := H.CronRepository.Update(H.CRD.Status.TestStatus.CronID, int(H.CRD.Spec.TestScheduleSeconds), H._getTestFunctionCron())
		if err != nil {
			return err, cron.EntryID(0)
		}
		H.CRD.Status.TestStatus.CronID = cronId.(cron.EntryID)
	}
	return nil, cron.EntryID(0)
}
