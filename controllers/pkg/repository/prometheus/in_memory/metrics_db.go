package in_memory

import (
	"fr.esgi/ha-audit/api/v1beta1"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
	"sync"
)

var MetricDB []DBEntry
var mutex = &sync.Mutex{}

type DBEntry struct {
	Id              string
	Name            string
	HistogramMetric prometheus.Histogram
	GaugeMetric     prometheus.Gauge
	CounterMetric   prometheus.Counter
}

func Create(name string, metric interface{}) string {
	var entry DBEntry
	id, _ := uuid.NewUUID()
	emptyGauge := prometheus.NewGauge(prometheus.GaugeOpts{})
	emptyHistogram := prometheus.NewHistogram(prometheus.HistogramOpts{})
	emptyCounter := prometheus.NewCounter(prometheus.CounterOpts{})
	if strings.ToLower(reflect.TypeOf(metric).String()) == strings.ToLower(reflect.TypeOf(&emptyGauge).String()) {
		entry = DBEntry{
			Id:          id.String(),
			Name:        name,
			GaugeMetric: metric.(prometheus.Gauge),
		}
	} else if strings.ToLower(reflect.TypeOf(metric).String()) == strings.ToLower(reflect.TypeOf(&emptyHistogram).String()) {
		entry = DBEntry{
			Id:              id.String(),
			Name:            name,
			HistogramMetric: metric.(prometheus.Histogram),
		}
	} else if strings.ToLower(reflect.TypeOf(metric).String()) == strings.ToLower(reflect.TypeOf(&emptyCounter).String()) {
		entry = DBEntry{
			Id:            id.String(),
			Name:          name,
			CounterMetric: metric.(prometheus.Counter),
		}
	}
	MetricDB = append(MetricDB, entry)
	return id.String()
}

func Update(id string, metric float64) {
	for _, e := range MetricDB {
		if e.Id == id {
			if e.GaugeMetric != nil {
				e.GaugeMetric.Set(metric)
			} else if e.HistogramMetric != nil {
				e.HistogramMetric.Observe(metric)
			} else if e.CounterMetric != nil {
				e.CounterMetric.Add(metric)
			}
		}
	}
}

func GetByID(id string) *DBEntry {
	for _, metric := range MetricDB {
		if metric.Id == id {
			return &metric
		}
	}
	return nil
}

func Get(metricType v1beta1.PrometheusMetricType, name string) *DBEntry {
	for _, metric := range MetricDB {
		switch metricType {
		case v1beta1.PrometheusMetricTypeGauge:
			if metric.GaugeMetric != nil && metric.Name == name {
				return &metric
			}
		case v1beta1.PrometheusMetricTypeHistogram:
			if metric.CounterMetric != nil && metric.Name == name {
				return &metric
			}
		case v1beta1.PrometheusMetricTypeCounter:
			if metric.HistogramMetric != nil && metric.Name == name {
				return &metric
			}
		}
	}
	return nil
}

func Delete(metricType v1beta1.PrometheusMetricType, name string) {
	for i, metric := range MetricDB {
		if metric.Name != name {
			break
		} else if metricType == v1beta1.PrometheusMetricTypeCounter && metric.CounterMetric != nil {
			MetricDB = append(MetricDB[:i], MetricDB[i+1:]...)
		} else if metricType == v1beta1.PrometheusMetricTypeGauge && metric.GaugeMetric != nil {
			MetricDB = append(MetricDB[:i], MetricDB[i+1:]...)
		} else if metricType == v1beta1.PrometheusMetricTypeHistogram && metric.HistogramMetric != nil {
			MetricDB = append(MetricDB[:i], MetricDB[i+1:]...)
		}
	}
}
