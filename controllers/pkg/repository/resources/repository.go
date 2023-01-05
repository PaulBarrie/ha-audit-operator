package resources

import (
	"context"
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	v1api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"reflect"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var (
	Config        = ctrl.GetConfigOrDie()
	DynamicConfig = dynamic.NewForConfigOrDie(Config)
)

type ResourceRepository struct {
	Client  client.Client
	Context context.Context
}

var resourceRepositoryInstance *ResourceRepository
var lock = &sync.Mutex{}

func GetInstance(client client.Client, ctx context.Context) *ResourceRepository {
	if resourceRepositoryInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if resourceRepositoryInstance == nil {
			resourceRepositoryInstance = &ResourceRepository{
				Client:  client,
				Context: ctx,
			}
		} else {
			kernel.Logger.Info("ResourceRepository already initialized")
		}
	} else {
		kernel.Logger.Info("ResourceRepository already initialized")
	}
	return resourceRepositoryInstance
}

func (r *ResourceRepository) Get(target v1beta1.Target) (TargetResourcePayload, error) {
	var pods []v1api.Pod
	var err error
	if target.Name != "" {
		pods, err = r._getResource(GetResourceTypePayload(target.Kind), target.Namespace, target.Name)
	} else if !reflect.DeepEqual(target.LabelSelector, map[string]string{}) {
		pods, err = r._getResource(GetResourceTypePayload(target.Kind), target.Namespace, target.LabelSelector)
	} else if target.NameRegex != "" {
		pods, err = r._getResource(GetResourceTypePayload(target.Kind), target.Namespace, target.NameRegex, true)
	} else {
		pods, err = r._getResource(GetResourceTypePayload(target.Kind), target.Namespace)
	}
	if err != nil {
		return TargetResourcePayload{}, err
	}
	return TargetResourcePayload{
		Id:         target.Id,
		TargetType: target.Kind,
		Pods:       pods,
	}, nil
}

func (r *ResourceRepository) GetAll(targets []v1beta1.Target) ([]TargetResourcePayload, error) {
	var res []TargetResourcePayload
	for _, target := range targets {
		targetResource, err := r.Get(target)
		if err != nil {
			kernel.Logger.Error(err, "error while getting resources")
			break
		}
		res = append(res, targetResource)
	}
	return res, nil
}

func (r *ResourceRepository) Delete(pod *v1api.Pod) error {
	err := r.Client.Delete(r.Context, pod)
	if err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("Error deleting pod %s in %s namespace", pod.GetName(), pod.GetNamespace()))
	}
	return nil
}

func (r *ResourceRepository) _getResource(resourcePayload ResourceTypePayload, namespace string, opts ...interface{}) (
	[]v1api.Pod, error) {
	var options metav1.ListOptions = metav1.ListOptions{}
	var res []v1api.Pod

	// Default: find in all the namespace
	// Find with name
	if len(opts) == 1 && reflect.TypeOf(opts[0]).Kind() == reflect.String {
		options = metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", opts[0]),
		}
	} else if len(opts) == 1 && reflect.TypeOf(opts[0]).Kind() == reflect.Map { // Find with label selector
		options = metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
				MatchLabels: opts[0].(map[string]string),
			}),
		}
	} else if len(opts) == 2 && opts[0] != nil && reflect.TypeOf(opts[0]).Kind() == reflect.String && opts[1].(bool) {
		podList, err := r.getPodsFromRegex(namespace, opts[0].(string))
		if err != nil {
			return []v1api.Pod{}, err
		}
		for _, pod := range podList {
			res = _appendOnce(res, pod)
		}
	}

	resourceId := schema.GroupVersionResource{
		Group:    resourcePayload.Group,
		Version:  resourcePayload.ApiVersion,
		Resource: resourcePayload.Kind,
	}

	list, err := DynamicConfig.Resource(resourceId).Namespace(namespace).
		List(r.Context, options)

	if err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("Cannot get resource %v with options %v in %s namespace", resourceId, options, namespace))
		return []v1api.Pod{}, err
	}
	res, err = r._toPodList(list)

	return res, nil
}

func (r *ResourceRepository) _toPodList(list *unstructured.UnstructuredList) ([]v1api.Pod, error) {
	var res []v1api.Pod
	for _, item := range list.Items {
		if item.GetKind() == "PodTarget" {
			var pod v1api.Pod
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &pod); err != nil {
				return []v1api.Pod{}, err
			}
			res = _appendOnce(res, pod)
		} else if item.GetKind() == "DeploymentTarget" || item.GetKind() == "StatefulSetTarget" ||
			item.GetKind() == "Daemonset" || item.GetKind() == "ReplicaSetTarget" {
			toAppend, err1 := r._getResource(
				GetResourceTypePayload("PodTarget"),
				item.GetNamespace(),
				fmt.Sprintf("%s-*", item.GetName()),
				true,
			)
			if err1 != nil {
				return []v1api.Pod{}, err1
			}
			for _, pod := range toAppend {
				res = append(res, pod)
			}
		}
	}
	return res, nil
}

func (r *ResourceRepository) getPodsFromRegex(namespace string, regex string) ([]v1api.Pod, error) {
	var res []v1api.Pod
	var resources v1api.PodList

	err := r.Client.List(r.Context, &resources, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return []v1api.Pod{}, err
	}

	for _, resource := range resources.Items {
		if match, _ := regexp.MatchString(regex, resource.GetName()); match {
			res = _appendOnce(res, resource)
		}
	}
	return res, nil
}

func _appendOnce(slice []v1api.Pod, i v1api.Pod) []v1api.Pod {
	for _, ele := range slice {
		if ele.GetName() == i.GetName() {
			return slice
		}
	}
	return append(slice, i)
}
