package prometheus

import (
	"fmt"
	"fr.esgi/ha-audit/api/v1beta1"
	metric_db "fr.esgi/ha-audit/controllers/pkg/db/metric"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sync"
)

const (
	DefaultJob = "ha-audit"
)

type Repository struct {
	Address    string `json:"address"`
	DefaultJob string `json:"defaultJob"`
}

var prometheusRepositoryInstance *Repository
var mutex = &sync.Mutex{}

func GetInstance(address string) *Repository {
	if prometheusRepositoryInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if prometheusRepositoryInstance == nil {
			prometheusRepositoryInstance = &Repository{
				Address:    address,
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

	switch prometheusMetric.Type {
	case v1beta1.PrometheusMetricTypeCounter:
		prometheusData := prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheusMetric.Name,
			Help: prometheusMetric.Help,
		})
		if err := metrics.Registry.Register(prometheusData); err != nil {
			kernel.Logger.Info(fmt.Sprintf("metric %s already exist", prometheusMetric.Name))
		}
		metric_db.Create(prometheusMetric.Name, prometheusData)
		kernel.Logger.Info(fmt.Sprintf("metric %s created", prometheusMetric.Name))
		return prometheusData, nil

	case v1beta1.PrometheusMetricTypeHistogram:
		prometheusData := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    prometheusMetric.Name,
			Help:    prometheusMetric.Help,
			Buckets: prometheus.DefBuckets,
		})
		if err := metrics.Registry.Register(prometheusData); err != nil {
			kernel.Logger.Info(fmt.Sprintf("metric %s already exist", prometheusMetric.Name))
		}
		metric_db.Create(prometheusMetric.Name, prometheusData)
		return prometheusData, nil
	case v1beta1.PrometheusMetricTypeGauge:
		prometheusData := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheusMetric.Name,
			Help: prometheusMetric.Help,
		})
		if err := metrics.Registry.Register(prometheusData); err != nil {
			kernel.Logger.Info(fmt.Sprintf("metric %s already exist", prometheusMetric.Name))
		}
		metric_db.Create(prometheusMetric.Name, prometheusData)
		return prometheusData, nil
	default:
		return nil, kernel.ErrorInvalidArgument(fmt.Sprintf("metric type %s is not supported", prometheusMetric.Type))
	}
}

func (p *Repository) Update(args ...interface{}) (interface{}, error) {
	if len(args) != 2 && args[0] != nil && reflect.TypeOf(args[1]).Kind() != reflect.Float64 {
		return nil, kernel.ErrorInvalidArgument("args must be a string and a float64")
	}
	metric := args[0]
	value := args[1].(float64)

	kernel.Logger.Info(fmt.Sprintf("%s metric %s updated, with value %f", args[0].(v1beta1.PrometheusMetric).Type,
		args[0].(v1beta1.PrometheusMetric).Name, args[1].(float64)))

	switch reflect.TypeOf(metric) {
	case reflect.TypeOf(prometheus.NewHistogram(prometheus.HistogramOpts{})):
		histogram := metric.(prometheus.Histogram)
		histogram.Observe(value)
	case reflect.TypeOf(prometheus.NewCounter(prometheus.CounterOpts{})):
		counter := metric.(prometheus.Counter)
		counter.Add(value)
	case reflect.TypeOf(prometheus.NewGauge(prometheus.GaugeOpts{})):
		gauge := metric.(prometheus.Gauge)
		gauge.Set(value)
	default:
		return nil, kernel.ErrorInvalidArgument(
			fmt.Sprintf(
				"First arg must be a prometheus metric and not : %s. Accepted: %s or %s",
				metric, v1beta1.PrometheusMetricTypeCounter, v1beta1.PrometheusMetricTypeHistogram,
			),
		)
	}
	return metric, nil
}

func (p *Repository) Delete(i interface{}) error {
	//TODO implement me
	panic("implement me")
}
