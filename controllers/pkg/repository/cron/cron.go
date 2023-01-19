package cron

type Payload struct {
	FrequencySec int    `json:"frequencySec"`
	Function     func() `json:"function"`
	CronId       int    `json:"cronId"`
}

type CronType string

const (
	CronChaosStrategy CronType = "chaosStrategy"
	CronReport        CronType = "report"
)
