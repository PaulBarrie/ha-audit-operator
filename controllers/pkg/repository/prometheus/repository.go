package prometheus

import (
	"fr.esgi/ha-audit/api/v1beta1"
	"fr.esgi/ha-audit/controllers/pkg/kernel"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"reflect"
	"sync"
)

const (
	DefaultJob = "ha-audit"
)

type PrometheusRepository struct {
	Address string `json:"address"`
	DefaultJob string `json:"defaultJob"`
}

var prometheusRepositoryInstance *PrometheusRepository
var mutex = &sync.Mutex{}

func GetInstance(address string) *PrometheusRepository {
	if prometheusRepositoryInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if prometheusRepositoryInstance == nil {
			prometheusRepositoryInstance = &PrometheusRepository{
				Address: address,
				DefaultJob: DefaultJob,
			}
		} else {
			kernel.Logger.Info("CronRepository already initialized")
		}
	} else {
		kernel.Logger.Info("CronRepository already initialized")
	}
	return prometheusRepositoryInstance
}

func (p *PrometheusRepository) Get(args ...interface{}) (interface{}, error) {
	return p._getOrCreate(args...)
}

func (p *PrometheusRepository) GetAll(i interface{}) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PrometheusRepository) Create(args ...interface{}) (interface{}, error) {
	return p._getOrCreate(args...)
}

func (p *PrometheusRepository) Update(args ...interface{}) error {
	if len(args) != 2 && reflect.TypeOf(args[0]) != reflect.TypeOf(v1beta1.PrometheusMetric{}) && reflect.TypeOf(args[1]).Kind() != reflect.Float64 {
		return kernel.ErrorInvalidArgument("args must be a string and a float64")
	}
	metric, err := p.Get(args[0])
	if err != nil {
		kernel.Logger.Error(err, "unable to get prometheus metric")
		return err
	}

	switch reflect.TypeOf(args[0]) {
	case reflect.TypeOf(prometheus.Histogram(nil)):
		histogram := metric.(prometheus.Histogram)
		histogram.Observe(args[1].(float64))
		prometheus.MustRegister(args[0].(prometheus.Histogram))
		if err = push.New(p.Address, p.DefaultJob).Collector(histogram).Push(); err != nil {
			kernel.Logger.Error(err, "unable to push prometheus metric")
			return err
		}
	case reflect.TypeOf(prometheus.Counter(nil)):
		counter := metric.(prometheus.Counter)
		counter.Add(args[1].(float64))
		prometheus.MustRegister(args[0].(prometheus.Counter))
		if err = push.New(p.Address, p.DefaultJob).Collector(counter).Push(); err != nil {
			kernel.Logger.Error(err, "unable to push prometheus metric")
			return err
		}
	default:
		return kernel.ErrorInvalidArgument("First arg must be a prometheus metric")
	}
	return nil
}

func (p *PrometheusRepository) Delete(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (p *PrometheusRepository) _getOrCreate(args ...interface{}) (interface{}, error) {
	var help = ""
	if len(args) >= 2 && reflect.TypeOf(args[0]).Kind() != reflect.String && reflect.TypeOf(args[1]).Kind() != reflect.String {
		return nil, kernel.ErrorInvalidArgument("args must be 2 string")
	}
	if len(args) == 3 && reflect.TypeOf(args[2]).Kind() != reflect.String {
		help = args[1].(string)
	}
	prometheusType := args[0].(v1beta1.PrometheusMetricType)
	name := args[1].(string)

	switch prometheusType {
	case v1beta1.PrometheusMetricTypeCounter:
		prometheusData := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
		prometheus.MustRegister(prometheusData)
		return prometheusData, nil
	case v1beta1.PrometheusMetricTypeRate:
		prometheusData := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: prometheus.DefBuckets,
		})
		prometheus.MustRegister(prometheusData)
		return prometheusData, nil
	}
	return nil, kernel.ErrorInvalidArgument("args must be 2 string and an optional one")
}
