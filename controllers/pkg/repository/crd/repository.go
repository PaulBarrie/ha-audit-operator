package crd

import (
	"context"
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type Repository struct {
	Client  client.Client
	Context context.Context
}

var crdRepositoryInstance *Repository
var mutex = &sync.Mutex{}

func GetInstance(client client.Client, ctx context.Context) *Repository {
	if crdRepositoryInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if crdRepositoryInstance == nil {
			crdRepositoryInstance = &Repository{
				Client:  client,
				Context: ctx,
			}
		}
	}
	return crdRepositoryInstance
}

func (r *Repository) Get(args ...interface{}) (interface{}, error) {
	if len(args) == 0 || reflect.TypeOf(args[0]).Kind() != reflect.String {
		return nil, kernel.ErrorInvalidArgument("arg must be CRD name")
	}
	namespace := kernel.DefaultNamespace
	if len(args) == 2 && reflect.TypeOf(args[1]).Kind() == reflect.String {
		namespace = args[1].(string)
	}
	crd := &v1beta1.HAAudit{}
	if err := r.Client.Get(r.Context, client.ObjectKey{Name: args[0].(string), Namespace: namespace}, crd); err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("unable to get CRD with name: %s", args[0].(string)))
		return nil, err
	}
	return crd, nil
}

func (r *Repository) GetAll(args interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Update(args ...interface{}) error {
	if len(args) != 1 || args[0] == nil || reflect.TypeOf(args[0]) != reflect.TypeOf(&v1beta1.HAAudit{}) {
		return kernel.ErrorInvalidArgument("args must be a HAAudit CRD")
	}
	newCrd := args[0].(*v1beta1.HAAudit)
	latestCRD := &v1beta1.HAAudit{}
	if err := r.Client.Get(r.Context, client.ObjectKey{Name: newCrd.Name, Namespace: newCrd.Namespace}, latestCRD); err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("unable to get CRD with name: %s", newCrd.Name))
		return err
	}
	latestCRD.Spec = newCrd.Spec
	latestCRD.Status = newCrd.Status
	//kernel.Logger.Info(fmt.Sprintf("NewCRD: %v", newCrd.Status))
	//kernel.Logger.Info(fmt.Sprintf("LatestCRD: %v", latestCRD.Status))
	//if err := r.Client.Update(r.Context, latestCRD); err != nil {
	//	kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD : %v", err))
	//}
	//if err := r.Client.Status().Update(r.Context, latestCRD); err != nil {
	//	kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD status : %v", err))
	//}
	//errRetry := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	//	kernel.Logger.Info(fmt.Sprintf("updating CRD with version: %s", newCrd.GetResourceVersion()))
	//	err := r.Client.Update(
	//		r.Context,
	//		newCrd)
	//	if err != nil {
	//		kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD status : %v", err))
	//	}
	//	return err
	//})
	//if errRetry != nil {
	//	return errRetry
	//}
	//statusErrRetry := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	//	kernel.Logger.Info(fmt.Sprintf("updating CRD status with version: %s", newCrd.GetResourceVersion()))
	//	err := r.Client.Status().Update(r.Context, newCrd)
	//	if err != nil {
	//		kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD status : %v", err))
	//	}
	//	return err
	//})
	//if statusErrRetry != nil {
	//	return statusErrRetry
	//}
	return nil
}

func (r *Repository) Create(i ...interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Delete(arg interface{}) error {
	if reflect.TypeOf(arg) != reflect.TypeOf(&v1beta1.HAAudit{}) {
		return kernel.ErrorInvalidArgument("arg must be CRD")
	}
	err := r.Client.Delete(r.Context, arg.(client.Object))
	if err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("unable to delete CRD with version: %s", arg.(client.Object).GetResourceVersion()))
		return err
	}
	return nil
}
