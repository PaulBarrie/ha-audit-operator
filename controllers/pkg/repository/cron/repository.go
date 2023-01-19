package cron

import (
	"fmt"
	cron_db "fr.esgi/ha-audit/controllers/pkg/db/cron"
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

func (c *Repository) Get(args ...interface{}) (interface{}, error) {
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

func (c *Repository) GetAll(i interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Repository) Create(args ...interface{}) (interface{}, error) {
	if len(args) != 2 && reflect.TypeOf(args[0]).Kind() != reflect.Int && reflect.TypeOf(args[1]).Kind() != reflect.Func {
		return nil, kernel.ErrorInvalidArgument("args must be a string and a func")
	}
	id, err := c.Cron.AddFunc(
		fmt.Sprintf(
			"@every %ds", args[0].(int)),
		args[1].(func()),
	)
	if err != nil {
		kernel.Logger.Error(err, "unable to add func to cron")
		return nil, err
	}
	c.Cron.Start()
	go c.Cron.Run()

	return id, nil
}

func (c *Repository) Update(args ...interface{}) (interface{}, error) {
	if len(args) != 2 &&
		reflect.TypeOf(args[0]).Kind() != reflect.Int &&
		reflect.TypeOf(args[1]) != reflect.TypeOf(Payload{}) {
		return args[0].(Payload).CronId, kernel.ErrorInvalidArgument("args must be the old cronID and new cron payloads")
	}
	newCron := args[1].(Payload)
	oldCronId := args[0].(int)

	c.Cron.Remove(cron.EntryID(oldCronId))
	cronId, err := c.Create(_getSeconds(newCron.FrequencySec), newCron.Function)
	if err != nil {
		kernel.Logger.Error(err, "unable to update cron")
		return cronId, err
	}
	return cronId, nil
}

func (c *Repository) Delete(args interface{}) error {
	if reflect.TypeOf(args) != reflect.TypeOf(cron.EntryID(0)) {
		return kernel.ErrorInvalidArgument("args must be an cron.EntryID")
	}
	cronID := args.(cron.EntryID)
	if cronID == 0 {
		kernel.Logger.Info("cron not found", "cronID", cronID)
		return nil
	}

	cron_db.Delete(cronID)
	return nil
}

func _getSeconds(seconds int) string {
	return fmt.Sprintf("@every %ds", seconds)
}
