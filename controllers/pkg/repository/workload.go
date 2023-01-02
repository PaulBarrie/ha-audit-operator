package repository

import (
	"context"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	v1 "k8s.io/api/apps/v1"
	v1api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadRepository struct {
	Workload Workload
	Client   client.Client
	Context  context.Context
}

type Workload struct {
	Resource interface{}
	Type     WorkloadResourceType
}

type WorkloadResourceType string

const (
	Deployment    WorkloadResourceType = "Deployment"
	Pod           WorkloadResourceType = "Pod"
	StatefulSet   WorkloadResourceType = "StatefulSet"
	DaemonSet     WorkloadResourceType = "DaemonSet"
	ReplicaSet    WorkloadResourceType = "ReplicaSet"
	ResourceError WorkloadResourceType = ""
)

func (w *WorkloadRepository) Get(name types.NamespacedName, resourceType WorkloadResourceType) Workload {
	switch resourceType {
	case Deployment:
		deployment := v1.Deployment{}
		if err := w.Client.Get(w.Context, name, &deployment); err != nil {
			kernel.Logger.Error(err, "unable to fetch Deployment")
			return Workload{Resource: nil, Type: ResourceError}
		}
		return Workload{Resource: deployment, Type: Deployment}
	case Pod:
		pod := v1api.Pod{}
		if err := w.Client.Get(w.Context, name, &pod); err != nil {
			kernel.Logger.Error(err, "unable to fetch Pod")
			return Workload{Resource: nil, Type: ResourceError}
		}
		return Workload{Resource: pod, Type: Pod}
	case StatefulSet:
		statefulSet := v1.StatefulSet{}
		if err := w.Client.Get(w.Context, name, &statefulSet); err != nil {
			kernel.Logger.Error(err, "unable to fetch StatefulSet")
			return Workload{Resource: nil, Type: ResourceError}
		}
		return Workload{Resource: statefulSet, Type: StatefulSet}
	case DaemonSet:
		daemonSet := v1.DaemonSet{}
		if err := w.Client.Get(w.Context, name, &daemonSet); err != nil {
			kernel.Logger.Error(err, "unable to fetch DaemonSet")
			return Workload{Resource: nil, Type: ResourceError}
		}
		return Workload{Resource: daemonSet, Type: DaemonSet}
	case ReplicaSet:
		replicaSet := v1.ReplicaSet{}
		if err := w.Client.Get(w.Context, name, &replicaSet); err != nil {
			kernel.Logger.Error(err, "unable to fetch ReplicaSet")
			return Workload{Resource: nil, Type: ResourceError}
		}
		return Workload{Resource: replicaSet, Type: ReplicaSet}
	}
	return Workload{Resource: nil, Type: ResourceError}
}

func (w *WorkloadRepository) GetAll(i interface{}) []interface{} {
	//TODO implement me
	panic("implement me")
}

func (w *WorkloadRepository) Delete() error {
	//TODO implement me
	panic("implement me")
}

func (p *Workload) New() error {

	return nil
}

/*
func (p *Workload) Get() (error, WorkloadResource, resourceType string) {
	if reflect.TypeOf(p) == reflect.TypeOf(v1.Deployment{}) {
		return nil, "", p
	} else if !reflect.DeepEqual(p.DaemonSet, v1.DaemonSet{}) {
		return nil, p.DaemonSet
	} else if !reflect.DeepEqual(p.StatefulSet, v1.StatefulSet{}) {
		return nil, p.StatefulSet
	} else if !reflect.DeepEqual(p.ReplicaSet, v1.ReplicaSet{}) {
		return nil, p.ReplicaSet
	} else if !reflect.DeepEqual(p.Pod, v1api.Pod{}) {
		return nil, p.Pod
	}
	return errors.New(fmt.Sprintf("The provided resource is not of the correct type (%s).\n"+
		"Should be either Pod, Deployment, Daemonset, Replicaset or Statefulset", reflect.TypeOf(p))), nil, ""
}
*/
/*
func (p *Workload) GetAll(targets []v1beta1.Target) []Workload {
	var resources []Workload
	for _, target := range targets {
		if target.Type == v1beta1.AuditTargetType("Pod") {
			resources = append(resources, Workload{getPod(target), WorkloadResourceType(target.Type)})
		} else if target.Type == v1beta1.AuditTargetType("Deployment") {
			resources = append(resources, Workload{getDeployment(target), WorkloadResourceType(target.Type)})
		} else if target.Type == v1beta1.AuditTargetType("Daemonset") {
			resources = append(resources, Workload{getDaemonset(target), WorkloadResourceType(target.Type)})
		} else if target.Type == v1beta1.AuditTargetType("Replicaset") {
			resources = append(resources, Workload{getReplicaset(target), WorkloadResourceType(target.Type)})
		} else if target.Type == v1beta1.AuditTargetType("Statefulset") {
			resources = append(resources, Workload{getStatefulset(target), WorkloadResourceType(target.Type)})
		} else {
			controllers.Logger.Info(
				fmt.Sprintf("%s is not a valid resource type", string(target.Type)),
			)
		}
	}
	return resources
}

func getDeployment(target v1beta1.Target) v1.Deployment {
	return v1.Deployment{}
}
func getReplicaset(target v1beta1.Target) interface{} {
	//TODO implement me
	panic("implement me")
}

func getDaemonset(target v1beta1.Target) interface{} {
	return nil
}

func getPod(target v1beta1.Target) v1api.Pod {
	return v1api.Pod{}
}

func getStatefulset(target v1beta1.Target) v1.StatefulSet {

	return v1.StatefulSet{}
}

func getFromRegex(target v1beta1.Target) []interface{} {
	return nil
}

func (p *Workload) Delete() error {
	//TODO implement me
	panic("implement me")
}
*/
