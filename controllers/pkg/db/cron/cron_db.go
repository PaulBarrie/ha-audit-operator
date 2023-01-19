package cron

import (
	"fr.esgi/ha-audit/api/v1beta1"
	cron_repo "fr.esgi/ha-audit/controllers/pkg/repository/cron"
	"github.com/robfig/cron/v3"
	"reflect"
)

var CronDB []CronDBEntry

type CronDBEntry struct {
	CronID                 cron.EntryID
	ResourceNamespacedName v1beta1.NamespacedName
	CronType               cron_repo.CronType
}

func Create(entry CronDBEntry) {
	for _, e := range CronDB {
		if reflect.DeepEqual(entry.ResourceNamespacedName, e.ResourceNamespacedName) && entry.CronType == e.CronType {
			return
		}
	}
	CronDB = append(CronDB, entry)
}

func Exists(id cron.EntryID) bool {
	for _, e := range CronDB {
		if e.CronID == id {
			return true
		}
	}
	return false
}

func Get(namespacedName v1beta1.NamespacedName, cronType cron_repo.CronType) *CronDBEntry {
	for _, e := range CronDB {
		if reflect.DeepEqual(namespacedName, e.ResourceNamespacedName) && cronType == e.CronType {
			return &e
		}
	}
	return nil
}

func Delete(namespacedName v1beta1.NamespacedName, cronType cron_repo.CronType) {
	for i, e := range CronDB {
		if reflect.DeepEqual(namespacedName, e.ResourceNamespacedName) && cronType == e.CronType {
			CronDB = append(CronDB[:i], CronDB[i+1:]...)
			return
		}
	}
}
