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

func (C Repository) Get(args ...interface{}) (interface{}, error) {
	panic("implement me")
}

func (C Repository) GetAll(args interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (C Repository) Update(args ...interface{}) error {
	if len(args) != 1 || args[0] == nil || reflect.TypeOf(args[0]) != reflect.TypeOf(&v1beta1.HAAudit{}) {
		return kernel.ErrorInvalidArgument("args must new CRD value")
	}
	err := C.Client.Update(C.Context, args[0].(client.Object))
	if err != nil {
		kernel.Logger.Error(err, fmt.Sprintf("unable to update CRD with version: %s", args[0].(client.Object).GetResourceVersion()))
	}
	return err
}

func (C Repository) Create(i ...interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (C Repository) Delete(arg interface{}) error {
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
