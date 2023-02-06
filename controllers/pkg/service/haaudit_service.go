package service

import (
	"context"
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	cron_repo "fr.esgi/ha-audit/controllers/pkg/repository/cron_cache"
	"fr.esgi/ha-audit/controllers/pkg/repository/prometheus"
	resource_repo "fr.esgi/ha-audit/controllers/pkg/repository/resources"
	v1api "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var (
	RunningEnvironment = "DEV"
	lock               = &sync.Mutex{}
)

func init() {
	envVar := os.Getenv("RunningEnvironment")
	if envVar == "PROD" || envVar == "DEV" {
		RunningEnvironment = envVar
	}
}

type PodTargets struct {
	Targets []v1beta1.Target
	Objects []v1api.Pod
}

type HAAuditService struct {
	Client               client.Client
	EventRecorder        record.EventRecorder
	Context              context.Context
	Targets              []resource_repo.TargetResourcePayload
	ResourceRepository   *resource_repo.ResourceRepository
	CronRepository       *cron_repo.Repository
	PrometheusRepository *prometheus.Repository
}

func New(client client.Client, ctx context.Context, eventRecorder record.EventRecorder) *HAAuditService {
	return &HAAuditService{
		Client:               client,
		Context:              ctx,
		EventRecorder:        eventRecorder,
		Targets:              []resource_repo.TargetResourcePayload{},
		ResourceRepository:   resource_repo.GetInstance(client, ctx),
		CronRepository:       cron_repo.GetInstance(),
		PrometheusRepository: prometheus.GetInstance(),
	}
}

func (H *HAAuditService) CreateOrUpdate(crd *v1beta1.HAAudit) error {
	if !crd.Status.Created {
		kernel.Logger.Info("Create HAAudit routines")
		return H.Create(crd)
	}
	//else {
	//	kernel.Logger.Info("Update HAAudit routines")
	//	err := H.Update(*crd)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (H *HAAuditService) Create(crd *v1beta1.HAAudit) error {
	newCRD := crd.DeepCopy()
	if RunningEnvironment == "DEV" {
		targets := H._inferTargets(crd)
		newCRD.Spec.Targets = targets
	}
	errStrat, cronID := H._scheduleStrategy(crd)
	if errStrat != nil {
		kernel.Logger.Error(errStrat, "unable to schedule strategy")
		return errStrat
	}
	newCRD.Status.ChaosStrategyCron = cronID
	payload := H.initPrometheusReport(crd)
	newCRD.Status.MetricStatus = v1beta1.MetricStatus{
		TotalUpMetricID: payload.IdTotal,
		RateUpMetricID:  payload.IdRate,
	}

	errReport, testCronID := H._scheduleTestReport(crd)
	if errReport != nil {
		kernel.Logger.Error(errReport, "unable to schedule tests")
		return errReport
	}
	newCRD.Status.TestStatus.CronID = testCronID
	newCRD.Status.Created = true

	if !reflect.DeepEqual(crd.Status, newCRD.Status) {
		crd.Status = newCRD.Status
		latestCRD := &v1beta1.HAAudit{}
		if err := H.Client.Get(H.Context, client.ObjectKey{Name: crd.Name, Namespace: crd.Namespace}, latestCRD); err != nil {
			kernel.Logger.Error(err, "unable to get CRD")
		}
		crd.ObjectMeta.ResourceVersion = latestCRD.ObjectMeta.ResourceVersion
		if errUpdate := H.Client.Status().Update(H.Context, crd); errUpdate != nil {
			kernel.Logger.Error(errUpdate, fmt.Sprintf("unable to update CRD status : %v", errUpdate))
			return errUpdate
		}
	}

	return nil
}

//func (H *HAAuditService) Update(newCRD v1beta1.HAAudit) error {
//	actualCRD := &v1beta1.HAAudit{}
//	if err := H.Client.Get(H.Context, client.ObjectKey{Name: newCRD.Name, Namespace: newCRD.Namespace}, actualCRD); err != nil {
//		kernel.Logger.Error(err, "unable to get CRD")
//	}
//
//	kernel.Logger.Info("ID metric tot: %s - ID metric rate: %s", crd)
//	nothingTodo := reflect.DeepEqual(crd.Spec, newCRD.Spec)
//	updateAll := !reflect.DeepEqual(crd.Spec.Targets, newCRD.Spec.Targets)
//	updateStrategy := !reflect.DeepEqual(crd.Spec.ChaosStrategy, newCRD.Spec.ChaosStrategy)
//	targetPathChanged := func() bool {
//		for _, target := range crd.Spec.Targets {
//			for _, newTarget := range newCRD.Spec.Targets {
//				if target.Name != newTarget.Name || target.Path != newTarget.Path {
//					return true
//				}
//			}
//		}
//		return false
//	}
//	updateTestReport := crd.Spec.Report.PrometheusReport.DumpFrequencySeconds != newCRD.Spec.Report.PrometheusReport.DumpFrequencySeconds || targetPathChanged()
//	if nothingTodo {
//		return nil
//	} else if updateAll {
//		if err := H._scheduleStrategy(); err != nil {
//			kernel.Logger.Error(err, "unable to schedule strategy")
//			return err
//		}
//		if err := H._scheduleTestReport(); err != nil {
//			kernel.Logger.Error(err, "unable to schedule tests")
//			return err
//		}
//	} else if updateStrategy {
//		if err := H._scheduleStrategy(); err != nil {
//			kernel.Logger.Error(err, "unable to schedule strategy")
//		}
//	} else if updateTestReport {
//		if err := H._scheduleTestReport(); err != nil {
//			kernel.Logger.Error(err, "unable to schedule tests")
//			return err
//		}
//	}
//	if !reflect.DeepEqual(crd.Spec, newCRD.Spec) {
//		if err := crdRepository.Update(crd.Spec); err != nil {
//			kernel.Logger.Error(err, "unable to update CRD status")
//			return err
//		}
//	}
//	if !reflect.DeepEqual(H.CRD.Status, newCRD.Status) {
//		if err := H.CRDRepository.Update(crd.Spec); err != nil {
//			kernel.Logger.Error(err, "unable to update CRD specs")
//			return err
//		}
//	}
//
//	return nil
//}

func (H *HAAuditService) Delete(crd *v1beta1.HAAudit) error {
	lock.Lock()
	defer lock.Unlock()
	kernel.Logger.Info("Delete HAAudit routines")
	err := H.CronRepository.Delete(crd.Status.ChaosStrategyCron)
	if err != nil {
		kernel.Logger.Error(err, "unable to delete cron")
		return err
	}
	if err = H.CronRepository.Delete(crd.Status.TestStatus.CronID); err != nil {
		kernel.Logger.Error(err, "unable to delete cron")
		return err
	}
	//if err = crdRepository.Delete(crd); err != nil {
	//	kernel.Logger.Error(err, "unable to delete CRD")
	//	return err
	//}
	return nil
}
