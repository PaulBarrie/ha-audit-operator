package cron

import (
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
		} else {
			kernel.Logger.Info("CronRepository already initialized")
		}
	} else {
		kernel.Logger.Info("CronRepository already initialized")
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
	if len(args) != 2 && reflect.TypeOf(args[0]).Kind() != reflect.String && reflect.TypeOf(args[1]).Kind() != reflect.Func {
		return nil, kernel.ErrorInvalidArgument("args must be a string and a func")
	}
	id, err := c.Cron.AddFunc(args[0].(string), args[1].(func()))
	if err != nil {
		kernel.Logger.Error(err, "unable to add func to cron")
		return nil, err
	}
	return id, nil
}

func (c *CronRepository) Delete(args interface{}) error {
	if reflect.TypeOf(args).Kind() != reflect.Int {
		return kernel.ErrorInvalidArgument("args must be an int")
	}

	c.Cron.Remove(cron.EntryID(args.(int)))
	return nil
}
