package cron_cache

import (
	"github.com/robfig/cron/v3"
	"reflect"
	"sync"
)

var CronDB DB
var mutex_db = &sync.Mutex{}

type DB struct {
	Table []Payload
}

func GetDB() *DB {
	if reflect.DeepEqual(CronDB, DB{}) {
		mutex_db.Lock()
		defer mutex_db.Unlock()
		if reflect.DeepEqual(CronDB, DB{}) {
			CronDB = DB{
				Table: make([]Payload, 0),
			}
		}
	}
	return &CronDB
}

func (db *DB) Create(entry Payload) {
	CronDB.Table = append(CronDB.Table, entry)
}

func (db *DB) Exists(id cron.EntryID) bool {
	for _, e := range CronDB.Table {
		if e.CronId == id {
			return true
		}
	}
	return false
}

func (db *DB) Get(cronId cron.EntryID) *Payload {
	for _, e := range CronDB.Table {
		if e.CronId == cronId {
			return &e
		}
	}
	return nil
}

func (db *DB) Update(oldID cron.EntryID, newEntry Payload) {
	db.Create(newEntry)
	db.Delete(oldID)
}

func (db *DB) Delete(id cron.EntryID) {
	for i, e := range CronDB.Table {
		if e.CronId == id {
			CronDB.Table = append(CronDB.Table[:i], CronDB.Table[i+1:]...)
			return
		}
	}
}
