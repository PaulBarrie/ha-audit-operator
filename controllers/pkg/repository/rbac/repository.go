package rbac

import (
	"context"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Repository struct {
	Client  client.Client
	Context context.Context
}

var RepositoryInstance *Repository

func GetInstance(client client.Client, ctx context.Context) *Repository {
	if RepositoryInstance == nil {
		RepositoryInstance = &Repository{
			Client:  client,
			Context: ctx,
		}
	}
	return RepositoryInstance
}

func (r *Repository) Get(args ...interface{}) (interface{}, error) {
	if len(args) != 1 && reflect.TypeOf(args[0]) != reflect.TypeOf(types.NamespacedName{}) {
		return nil, kernel.ErrorInvalidArgument("arg must be a namespaced name")
	}
	var clusterRoleBind v1.ClusterRoleBinding
	if err := r.Client.Get(r.Context, client.ObjectKey{
		Name:      args[0].(string),
		Namespace: args[1].(string),
	}, &clusterRoleBind); err != nil {
		kernel.Logger.Error(err, "unable to get cluster role binding")
		return nil, err
	}
	return clusterRoleBind, nil
}

func (r *Repository) GetAll(arg interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Update(args ...interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Create(args ...interface{}) (interface{}, error) {
	if len(args) != 2 || reflect.TypeOf(args[0]) != reflect.TypeOf(v1beta1.ServiceAccount{}) {
		return nil, kernel.ErrorInvalidArgument("args must be a non empty string")
	}
	sa := args[0].(v1beta1.ServiceAccount)
	if sa.SAName == "" {
		return nil, kernel.ErrorInvalidArgument("args must be a non empty string")
	}
	if sa.SANamespace == "" {
		sa.SANamespace = kernel.DefaultNamespace
	}
	clusterRole := v1.ClusterRoleBinding{
		ObjectMeta: controllerruntime.ObjectMeta{
			Name: "prometheus-k8s-rolebinding",
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "prometheus-k8s-role",
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.SAName,
				Namespace: sa.SANamespace,
			},
		},
	}
	err := r.Client.Create(r.Context, &clusterRole)
	if err != nil {
		return nil, err
	}
	return clusterRole, nil
}

func (r *Repository) Delete(arg interface{}) error {
	if reflect.TypeOf(arg) != reflect.TypeOf(types.NamespacedName{}) {
		return kernel.ErrorInvalidArgument("arg must be a namespaced name")
	}
	clusterRoleBind, err := r.Get(arg.(types.NamespacedName))
	if err != nil {
		return err
	}
	err = r.Client.Delete(r.Context, clusterRoleBind.(*v1.ClusterRoleBinding))
	if err != nil {
		return err
	}
	return nil
}
