package prometheus

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	metric_db "fr.esgi/ha-audit/controllers/pkg/repository/prometheus/in_memory"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sync"
)

const (
	DefaultJob = "ha-audit"
)

type Repository struct {
	DefaultJob string `json:"defaultJob"`
}

var prometheusRepositoryInstance *Repository
var mutex = &sync.Mutex{}

func GetInstance() *Repository {
	if prometheusRepositoryInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if prometheusRepositoryInstance == nil {
			prometheusRepositoryInstance = &Repository{
				DefaultJob: DefaultJob,
			}
		}
	}
	return prometheusRepositoryInstance
}

func (p *Repository) Get(args ...interface{}) (interface{}, error) {
	panic("implement me")
}

func (p *Repository) GetAll(i interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Repository) Create(args ...interface{}) (interface{}, error) {
	if len(args) != 1 && reflect.TypeOf(args) != reflect.TypeOf(v1beta1.PrometheusMetric{}) {
		return nil, kernel.ErrorInvalidArgument("args must be a PrometheusMetric")
	}
	prometheusMetric := args[0].(v1beta1.PrometheusMetric)

	kernel.Logger.Info(fmt.Sprintf("metric %s creation", prometheusMetric.Name))
	switch prometheusMetric.Type {
	case v1beta1.PrometheusMetricTypeCounter:
		prometheusData := prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheusMetric.Name,
			Help: prometheusMetric.Help,
		})
		if err := prometheus.Register(prometheusData); err != nil {
			kernel.Logger.Info(fmt.Sprintf("metric %s already exist", prometheusMetric.Name))
		}
		id := metric_db.Create(prometheusMetric.Name, prometheusData)
		kernel.Logger.Info(fmt.Sprintf("Created with id : %s", id))
		kernel.Logger.Info(fmt.Sprintf("metric %s created", prometheusMetric.Name))
		return id, nil

	case v1beta1.PrometheusMetricTypeHistogram:
		prometheusData := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    prometheusMetric.Name,
			Help:    prometheusMetric.Help,
			Buckets: prometheus.DefBuckets,
		})
		if err := prometheus.Register(prometheusData); err != nil {
			kernel.Logger.Info(fmt.Sprintf("metric %s already exist", prometheusMetric.Name))
		}

		id := metric_db.Create(prometheusMetric.Name, prometheusData)
		kernel.Logger.Info(fmt.Sprintf("Created with id : %s", id))

		return id, nil
	case v1beta1.PrometheusMetricTypeGauge:
		prometheusData := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheusMetric.Name,
			Help: prometheusMetric.Help,
		})
		if err := prometheus.Register(prometheusData); err != nil {
			kernel.Logger.Info(fmt.Sprintf("metric %s already exist", prometheusMetric.Name))
		}

		id := metric_db.Create(prometheusMetric.Name, prometheusData)
		kernel.Logger.Info(fmt.Sprintf("Created with id : %s", id))
		return id, nil
	default:
		return nil, kernel.ErrorInvalidArgument(fmt.Sprintf("metric type %s is not supported", prometheusMetric.Type))
	}
}

func (p *Repository) Update(args ...interface{}) error {
	if len(args) != 2 && reflect.TypeOf(args[0]).Kind() == reflect.String && reflect.TypeOf(args[1]).Kind() != reflect.Float64 {
		return kernel.ErrorInvalidArgument("args must be a string and a float64")
	}
	metricId := args[0].(string)
	value := args[1].(float64)

	kernel.Logger.Info(fmt.Sprintf("metric %s updated with value %f", args[0].(string), args[1].(float64)))
	metric_db.Update(metricId, value)
	return nil
}

func (p *Repository) Delete(i interface{}) error {
	//TODO implement me
	panic("implement me")
}
