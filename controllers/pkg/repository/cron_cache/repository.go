package cron_cache

import (
	"fmt"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"reflect"
	"sync"
)

type Repository struct {
	Cron *cron.Cron
}
type InMemoryPrometheusRepository struct {
	Id    string
	Gauge prometheus.Gauge
}

var InMemoryPrometheusRepositoryInstance []InMemoryPrometheusRepository

var cronRepositoryInstance *Repository
var mutex = &sync.Mutex{}

func GetInstance() *Repository {
	if cronRepositoryInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if cronRepositoryInstance == nil {
			cronRepositoryInstance = &Repository{
				Cron: cron.New(),
			}
		}
	}
	return cronRepositoryInstance
}

func (r *Repository) Get(args ...interface{}) (interface{}, error) {
	if len(args) != 0 || reflect.TypeOf(args).Kind() != reflect.String {
		return nil, kernel.ErrorInvalidArgument("the argument must be a string")
	}
	for _, metric := range InMemoryPrometheusRepositoryInstance {
		if metric.Id == args[0].(string) {
			return metric, nil
		}
	}
	return nil, kernel.ErrorNotFound(fmt.Sprintf("No metric found with id : %s", args[0].(string)))
}

func (r *Repository) GetAll(i interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Repository) Create(args ...interface{}) (interface{}, error) {
	if len(args) != 2 && reflect.TypeOf(args[0]).Kind() != reflect.Int && reflect.TypeOf(args[1]).Kind() != reflect.Func {
		return nil, kernel.ErrorInvalidArgument("args must be a string and a func")
	}
	id, err := r.Cron.AddFunc(
		fmt.Sprintf(
			"@every %ds", args[0].(int)),
		args[1].(func()),
	)
	if err != nil {
		kernel.Logger.Error(err, "unable to add func to cron")
		return nil, err
	}
	r.Cron.Start()
	go r.Cron.Run()

	return id, nil
}

func (r *Repository) Update(args ...interface{}) (interface{}, error) {
	if len(args) != 2 &&
		reflect.TypeOf(args[0]).Kind() != reflect.Int &&
		reflect.TypeOf(args[1]) != reflect.TypeOf(Payload{}) {
		return args[0].(Payload).CronId, kernel.ErrorInvalidArgument("args must be the old cronID and new cron payloads")
	}
	newCron := args[1].(Payload)
	oldCronId := args[0].(int)

	r.Cron.Remove(cron.EntryID(oldCronId))
	cronId, err := r.Create(_getSeconds(newCron.FrequencySec), newCron.Function)
	if err != nil {
		kernel.Logger.Error(err, "unable to update cron")
		return cronId, err
	}
	r.Cron.Start()
	go r.Cron.Run()
	GetDB().Update(cron.EntryID(oldCronId), newCron)
	return cronId, nil
}

func (r *Repository) Delete(args interface{}) error {
	if reflect.TypeOf(args) != reflect.TypeOf(cron.EntryID(0)) {
		return kernel.ErrorInvalidArgument("args must be an cron.EntryID")
	}
	cronID := args.(cron.EntryID)
	if cronID == 0 {
		kernel.Logger.Info("cron not found", "cronID", cronID)
		return nil
	}
	return nil
}

func _getSeconds(seconds int) string {
	return fmt.Sprintf("@every %ds", seconds)
}
