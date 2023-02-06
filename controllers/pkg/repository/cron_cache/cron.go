package cron_cache

import "github.com/robfig/cron/v3"

type Payload struct {
	FrequencySec int          `json:"frequencySec"`
	Function     func()       `json:"function"`
	CronId       cron.EntryID `json:"cronId"`
}

type CronType string
