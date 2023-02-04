package service

import (
	"context"
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	crd_repo "fr.esgi/ha-audit/controllers/pkg/repository/crd"
	cron_repo "fr.esgi/ha-audit/controllers/pkg/repository/cron_cache"
	"fr.esgi/ha-audit/controllers/pkg/repository/prometheus"
	resource_repo "fr.esgi/ha-audit/controllers/pkg/repository/resources"
	v1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
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
	CRD                  *v1beta1.HAAudit
	Client               client.Client
	Context              context.Context
	Targets              []resource_repo.TargetResourcePayload
	ResourceRepository   *resource_repo.ResourceRepository
	CronRepository       *cron_repo.Repository
	PrometheusRepository *prometheus.Repository
	CRDRepository        *crd_repo.Repository
}

func New(client client.Client, ctx context.Context, crd *v1beta1.HAAudit) *HAAuditService {
	return &HAAuditService{
		CRD:                  crd,
		Client:               client,
		Context:              ctx,
		Targets:              []resource_repo.TargetResourcePayload{},
		ResourceRepository:   resource_repo.GetInstance(client, ctx),
		CronRepository:       cron_repo.GetInstance(),
		PrometheusRepository: prometheus.GetInstance(),
		CRDRepository:        crd_repo.GetInstance(client, ctx),
	}
}

func (H *HAAuditService) CreateOrUpdate() (error, *v1beta1.HAAudit) {
	kernel.Logger.Info(fmt.Sprintf("CRD art creation : %v", H.CRD.Status))

	if !H.CRD.Status.Created {
		kernel.Logger.Info("Create HAAudit routines")
		return H.Create()
		//if err != nil {
		//	return err, nil
		//}
	}
	endCRD := &v1beta1.HAAudit{}
	if err := H.Client.Get(H.Context, types.NamespacedName{Name: H.CRD.Name, Namespace: H.CRD.Namespace}, endCRD); err != nil {
		kernel.Logger.Error(err, "unable to get CRD")
		return err, nil
	}
	kernel.Logger.Info(fmt.Sprintf("CRD at end : %v", endCRD.Status))
	//else {
	//	kernel.Logger.Info("Update HAAudit routines")
	//	err := H.Update(*H.CRD)
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil, nil
}

func (H *HAAuditService) Create() (error, *v1beta1.HAAudit) {
	newCRD := H.CRD.DeepCopy()
	if RunningEnvironment == "DEV" {
		targets := H._inferTargets()
		newCRD.Spec.Targets = targets
	}
	err, cronID := H._scheduleStrategy()
	if err != nil {
		kernel.Logger.Error(err, "unable to schedule strategy")
		return err, nil
	}
	newCRD.Status.ChaosStrategyCron = cronID
	payload := H.initPrometheusReport()
	newCRD.Status.MetricStatus = v1beta1.MetricStatus{
		TotalUpMetricID: payload.IdTotal,
		RateUpMetricID:  payload.IdRate,
	}
	newCRD.Spec.Report.PrometheusReport = payload.Report
	err, testCronID := H._scheduleTestReport()
	if err != nil {
		kernel.Logger.Error(err, "unable to schedule tests")
		return err, nil
	}
	newCRD.Status.TestStatus.CronID = testCronID
	newCRD.Status.Created = true
	H.CRD.Status = newCRD.Status
	H.CRD.Spec = newCRD.Spec
	kernel.Logger.Info(fmt.Sprintf("CRD status: %v", H.CRD.Status))
	if err = H.CRDRepository.Update(newCRD); err != nil {
		kernel.Logger.Error(err, "unable to update CRD")
		return err, nil
	}
	//if err = H.Client.Update(H.Context, newCRD); err != nil {
	//	kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD : %v", err))
	//}
	//check := v1beta1.HAAudit{}
	//if err = H.Client.Get(H.Context, types.NamespacedName{Name: H.CRD.Name, Namespace: H.CRD.Namespace}, &check); err != nil {
	//	kernel.Logger.Error(err, "unable to get CRD")
	//	return err
	//}
	//kernel.Logger.Info(fmt.Sprintf("CRD modified: %t", check.Status.Created))
	return nil, H.CRD
}

//func (H *HAAuditService) Update(newCRD v1beta1.HAAudit) error {
//	actualCRD := &v1beta1.HAAudit{}
//	if err := H.Client.Get(H.Context, client.ObjectKey{Name: newCRD.Name, Namespace: newCRD.Namespace}, actualCRD); err != nil {
//		kernel.Logger.Error(err, "unable to get CRD")
//	}
//
//	kernel.Logger.Info("ID metric tot: %s - ID metric rate: %s", H.CRD)
//	nothingTodo := reflect.DeepEqual(H.CRD.Spec, newCRD.Spec)
//	updateAll := !reflect.DeepEqual(H.CRD.Spec.Targets, newCRD.Spec.Targets)
//	updateStrategy := !reflect.DeepEqual(H.CRD.Spec.ChaosStrategy, newCRD.Spec.ChaosStrategy)
//	targetPathChanged := func() bool {
//		for _, target := range H.CRD.Spec.Targets {
//			for _, newTarget := range newCRD.Spec.Targets {
//				if target.Name != newTarget.Name || target.Path != newTarget.Path {
//					return true
//				}
//			}
//		}
//		return false
//	}
//	updateTestReport := H.CRD.Spec.Report.PrometheusReport.DumpFrequencySeconds != newCRD.Spec.Report.PrometheusReport.DumpFrequencySeconds || targetPathChanged()
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
//	if !reflect.DeepEqual(H.CRD.Spec, newCRD.Spec) {
//		if err := H.CRDRepository.Update(H.CRD.Spec); err != nil {
//			kernel.Logger.Error(err, "unable to update CRD status")
//			return err
//		}
//	}
//	if !reflect.DeepEqual(H.CRD.Status, newCRD.Status) {
//		if err := H.CRDRepository.Update(H.CRD.Spec); err != nil {
//			kernel.Logger.Error(err, "unable to update CRD specs")
//			return err
//		}
//	}
//
//	return nil
//}

func (H *HAAuditService) Delete() error {
	lock.Lock()
	defer lock.Unlock()
	kernel.Logger.Info("Delete HAAudit routines")
	err := H.CronRepository.Delete(H.CRD.Status.ChaosStrategyCron)
	if err != nil {
		kernel.Logger.Error(err, "unable to delete cron")
		return err
	}
	if err = H.CronRepository.Delete(H.CRD.Status.TestStatus.CronID); err != nil {
		kernel.Logger.Error(err, "unable to delete cron")
		return err
	}
	if err = H.CRDRepository.Delete(H.CRD); err != nil {
		kernel.Logger.Error(err, "unable to delete CRD")
		return err
	}
	return nil
}
