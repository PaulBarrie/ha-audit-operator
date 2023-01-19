package metric

import (
	"fr.esgi/ha-audit/api/v1beta1"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
)

var MetricDB []DBEntry

type DBEntry struct {
	Id              string
	Name            string
	HistogramMetric *prometheus.Histogram
	GaugeMetric     *prometheus.Gauge
	CounterMetric   *prometheus.Counter
}

func Create(name string, metric interface{}) {
	var entry DBEntry
	id, _ := uuid.NewUUID()
	for _, e := range MetricDB {
		if e.Name == name {
			return
		}
		if reflect.TypeOf(metric) == reflect.TypeOf(prometheus.NewGauge(prometheus.GaugeOpts{})) {
			entry = DBEntry{
				Id:          id.String(),
				Name:        name,
				GaugeMetric: metric.(*prometheus.Gauge),
			}
		} else if reflect.TypeOf(metric) == reflect.TypeOf(prometheus.NewHistogram(prometheus.HistogramOpts{})) {
			entry = DBEntry{
				Id:              id.String(),
				Name:            name,
				HistogramMetric: metric.(*prometheus.Histogram),
			}
		} else if reflect.TypeOf(metric) == reflect.TypeOf(prometheus.NewCounter(prometheus.CounterOpts{})) {
			entry = DBEntry{
				Id:            id.String(),
				Name:          name,
				CounterMetric: metric.(*prometheus.Counter),
			}
		}
	}
	MetricDB = append(MetricDB, entry)
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
