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

func (C *Repository) Get(args ...interface{}) (interface{}, error) {
	if len(args) == 0 || reflect.TypeOf(args[0]).Kind() != reflect.String {
		return nil, kernel.ErrorInvalidArgument("arg must be CRD name")
	}
	namespace := kernel.DefaultNamespace
	if len(args) == 2 && reflect.TypeOf(args[1]).Kind() == reflect.String {
		namespace = args[1].(string)
	}
	crd := &v1beta1.HAAudit{}
	if err := C.Client.Get(C.Context, client.ObjectKey{Name: args[0].(string), Namespace: namespace}, crd); err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("unable to get CRD with name: %s", args[0].(string)))
		return nil, err
	}
	return crd, nil
}

func (C *Repository) GetAll(args interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (C *Repository) Update(args ...interface{}) error {
	if len(args) != 1 || args[0] == nil || reflect.TypeOf(args[0]) != reflect.TypeOf(&v1beta1.HAAudit{}) {
		return kernel.ErrorInvalidArgument("args must new CRD value")
	}
	newCrd := args[0].(*v1beta1.HAAudit)
	getCRD, err := C.Get(newCrd.Name, newCrd.Namespace)
	if err != nil {
		return err
	}
	actualCRD := getCRD.(*v1beta1.HAAudit)
	if !reflect.DeepEqual(actualCRD.Spec, newCrd.Spec) {
		if err = C.Client.Update(C.Context, newCrd); err != nil {
			kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD status : %v", err))
		}
	}
	if !reflect.DeepEqual(actualCRD.Status, newCrd.Status) {
		actualCRD.Status = newCrd.Status
		if err = C.Client.Status().Update(C.Context, actualCRD); err != nil {
			kernel.Logger.Info(fmt.Sprintf("unable to update CRD status : %v", err))
			return err
		}
	}
	return err
}

func (C *Repository) Create(i ...interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (C *Repository) Delete(arg interface{}) error {
	if reflect.TypeOf(arg) != reflect.TypeOf(&v1beta1.HAAudit{}) {
		return kernel.ErrorInvalidArgument("arg must be CRD")
	}
	err := C.Client.Delete(C.Context, arg.(client.Object))
	if err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("unable to delete CRD with version: %s", arg.(client.Object).GetResourceVersion()))
		return err
	}
	return nil
}
