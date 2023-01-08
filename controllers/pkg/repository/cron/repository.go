package cron

import (
	"fmt"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"github.com/robfig/cron/v3"
	"reflect"
	"sync"
)

type CronRepository struct {
	Cron *cron.Cron
}

var cronRepositoryInstance *CronRepository
var mutex = &sync.Mutex{}

func GetInstance() *CronRepository {
	if cronRepositoryInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if cronRepositoryInstance == nil {
			cronRepositoryInstance = &CronRepository{
				Cron: cron.New(),
			}
		}
	}
	return cronRepositoryInstance
}

func (c *CronRepository) Get(args ...interface{}) (interface{}, error) {
	panic("implement me")
}

func (c *CronRepository) GetAll(i interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CronRepository) Create(args ...interface{}) (interface{}, error) {
	if len(args) != 2 && reflect.TypeOf(args[0]).Kind() != reflect.Int && reflect.TypeOf(args[1]).Kind() != reflect.Func {
		return nil, kernel.ErrorInvalidArgument("args must be a string and a func")
	}
	id, err := c.Cron.AddFunc(
		fmt.Sprintf(
			"@every %ds", args[0].(int)),
		args[1].(func()),
	)
	kernel.Logger.Info("cron created", "cronID", id)
	if err != nil {
		kernel.Logger.Error(err, "unable to add func to cron")
		return nil, err
	}
	c.Cron.Start()
	go c.Cron.Run()

	return id, nil
}

func (c *CronRepository) Update(args ...interface{}) (interface{}, error) {
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

func (c *CronRepository) Delete(args interface{}) error {
	if reflect.TypeOf(args) != reflect.TypeOf(cron.EntryID(0)) {
		return kernel.ErrorInvalidArgument("args must be an cron.EntryID")
	}
	cronID := args.(cron.EntryID)
	if cronID == 0 {
		kernel.Logger.Info("cron not found", "cronID", cronID)
		return nil
	}

	c.Cron.Remove(cronID)
	return nil
}

func _getSeconds(seconds int) string {
	return fmt.Sprintf("@every %ds", seconds)
}
