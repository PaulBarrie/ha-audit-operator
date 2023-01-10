package service

import (
	"context"
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	crd_repo "fr.esgi/ha-audit/controllers/pkg/repository/crd"
	cron_repo "fr.esgi/ha-audit/controllers/pkg/repository/cron"
	"fr.esgi/ha-audit/controllers/pkg/repository/prometheus"
	rbac_repo "fr.esgi/ha-audit/controllers/pkg/repository/rbac"
	resource_repo "fr.esgi/ha-audit/controllers/pkg/repository/resources"
	v1api "k8s.io/api/core/v1"
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
	CRD                  *v1beta1.HAAudit
	Client               client.Client
	Context              context.Context
	Targets              []resource_repo.TargetResourcePayload
	ResourceRepository   *resource_repo.ResourceRepository
	CronRepository       *cron_repo.Repository
	PrometheusRepository *prometheus.Repository
	CRDRepository        *crd_repo.Repository
	RBACRepository       *rbac_repo.Repository
}

func New(client client.Client, ctx context.Context, crd *v1beta1.HAAudit) *HAAuditService {
	return &HAAuditService{
		CRD:                  crd,
		Client:               client,
		Context:              ctx,
		Targets:              []resource_repo.TargetResourcePayload{},
		ResourceRepository:   resource_repo.GetInstance(client, ctx),
		CronRepository:       cron_repo.GetInstance(),
		PrometheusRepository: prometheus.GetInstance(crd.Spec.Report.PrometheusReport.Address),
		CRDRepository:        crd_repo.GetInstance(client, ctx),
		RBACRepository:       rbac_repo.GetInstance(client, ctx),
	}
}

func (H *HAAuditService) CreateOrUpdate() error {

	if H.CRD.Status.StrategyCronId == 0 && H.CRD.Status.TestReportCronId == 0 {
		kernel.Logger.Info("Create HAAudit routines")
		err := H.Create()
		if err != nil {
			return err
		}
	} else {
		kernel.Logger.Info("Update HAAudit routines")
		err := H.Update(*H.CRD)
		if err != nil {
			return err
		}
	}
	return nil
}

func (H *HAAuditService) Create() error {
	if RunningEnvironment == "DEV" {
		H.CRD.Spec.Targets = H._inferTargets()
	}
	if err := H._scheduleStrategy(); err != nil {
		kernel.Logger.Error(err, "unable to schedule strategy")
		return err
	}

	if err := H._scheduleTestReport(); err != nil {
		kernel.Logger.Error(err, "unable to schedule tests")
		return err
	}
	if H.CRD.Spec.Report.PrometheusReport.Address != "" {
		H.initPrometheusReport()
	}
	if err := H.CRDRepository.Update(H.CRD); err != nil {
		kernel.Logger.Error(err, "unable to update CRD")
		return err
	}
	return nil
}

func (H *HAAuditService) Update(newCRD v1beta1.HAAudit) error {
	nothingTodo := reflect.DeepEqual(H.CRD.Spec, newCRD.Spec)
	updateAll := !reflect.DeepEqual(H.CRD.Spec.Targets, newCRD.Spec.Targets)
	updateStrategy := !reflect.DeepEqual(H.CRD.Spec.ChaosStrategy, newCRD.Spec.ChaosStrategy)
	targetPathChanged := func() bool {
		for _, target := range H.CRD.Spec.Targets {
			for _, newTarget := range newCRD.Spec.Targets {
				if target.Name != newTarget.Name || target.Path != newTarget.Path {
					return true
				}
			}
		}
		return false
	}
	updateTestReport := !reflect.DeepEqual(H.CRD.Spec.Report, newCRD.Spec.Report) || targetPathChanged()
	if nothingTodo {
		return nil
	} else if updateAll {
		if err := H._scheduleStrategy(); err != nil {
			kernel.Logger.Error(err, "unable to schedule strategy")
			return err
		}
		kernel.Logger.Info(fmt.Sprintf("11/CRD Version: %s", H.CRD.ObjectMeta.ResourceVersion))
		if err := H._scheduleTestReport(); err != nil {
			kernel.Logger.Error(err, "unable to schedule tests")
			return err
		}
		kernel.Logger.Info(fmt.Sprintf("12/CRD Version: %s", H.CRD.ObjectMeta.ResourceVersion))

	} else if updateStrategy {
		if err := H._scheduleStrategy(); err != nil {
			kernel.Logger.Error(err, "unable to schedule strategy")
		}
		kernel.Logger.Info(fmt.Sprintf("13/CRD Version: %s", H.CRD.ObjectMeta.ResourceVersion))
	} else if updateTestReport {
		if err := H._scheduleTestReport(); err != nil {
			kernel.Logger.Error(err, "unable to schedule tests")
			return err
		}
		kernel.Logger.Info(fmt.Sprintf("14/CRD Version: %s", H.CRD.ObjectMeta.ResourceVersion))

	}

	if err := H.CRDRepository.Update(H.CRD); err != nil {
		kernel.Logger.Error(err, "unable to update CRD")
		return err
	}

	return nil
}

func (H *HAAuditService) Delete() error {
	lock.Lock()
	defer lock.Unlock()
	kernel.Logger.Info("Delete HAAudit routines")
	err := H.CronRepository.Delete(H.CRD.Status.TestReportCronId)
	if err != nil {
		kernel.Logger.Error(err, "unable to delete cron")
		return err
	}
	if err = H.CronRepository.Delete(H.CRD.Status.StrategyCronId); err != nil {
		kernel.Logger.Error(err, "unable to delete cron")
		return err
	}
	//if err = H.RBACRepository.Delete(H.CRD.Status.PrometheusClusterRoleBinding); err != nil {
	//	kernel.Logger.Error(err, "unable to delete CRD")
	//	return err
	//}
	if err = H.CRDRepository.Delete(H.CRD); err != nil {
		kernel.Logger.Error(err, "unable to delete CRD")
		return err
	}
	return nil
}
