package v1beta1

import (
	"fmt"
	"strings"
)

type HAReport struct {
	// +kubebuilder:validation:Required
	PrometheusReport PrometheusReport `json:"prometheus"`
}

type PrometheusMetricType string

const (
	DefaultEventType              string               = "Normal"
	PrometheusMetricTypeCounter   PrometheusMetricType = "counter"
	PrometheusMetricTypeHistogram PrometheusMetricType = "histogram"
	PrometheusMetricTypeGauge     PrometheusMetricType = "gauge"
)

type MetricEventPayload struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type PrometheusMetric struct {
	Name               string               `json:"name"`
	Help               string               `json:"help"`
	Type               PrometheusMetricType `json:"type"`
	MetricEventPayload MetricEventPayload   `json:"metricEvenPayload"`
}

type PrometheusReport struct {
	// +kubebuilder:default:=10
	DumpFrequencySeconds int `json:"dumpFrequency"`
}

func DefaultTotalRunningInstanceMetric(name string, rate int) *PrometheusMetric {
	metricName := fmt.Sprintf("%s_total_running_instance", strings.ReplaceAll(name, "-", "_"))
	return &PrometheusMetric{
		Type: PrometheusMetricTypeGauge,
		Name: metricName,
		Help: fmt.Sprintf(
			"Total number of running instance for %s report recorded every %d seconds",
			name,
			rate,
		),
		MetricEventPayload: MetricEventPayload{
			Type:    DefaultEventType,
			Reason:  "MetricCreated",
			Message: fmt.Sprintf("Total number of running instance recorded in metric : %s", metricName),
		},
	}
}

func DefaultTotalRunningInstanceRateMetric(name string, rate int) *PrometheusMetric {
	metricName := fmt.Sprintf("%s_rate_running_instance_seconds", strings.ReplaceAll(name, "-", "_"))
	return &PrometheusMetric{
		Type: PrometheusMetricTypeGauge,
		Name: metricName,
		Help: fmt.Sprintf(
			"The running instance rate for %s report recorded every %d seconds",
			name,
			rate,
		),
		MetricEventPayload: MetricEventPayload{
			Type:    DefaultEventType,
			Reason:  "MetricCreated",
			Message: fmt.Sprintf("Rate of running instance recorded in metric : %s", metricName),
		},
	}
}
