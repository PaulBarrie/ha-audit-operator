package v1beta1

import (
	"fmt"
	"reflect"
	"strings"
)

type HAReport struct {
	// +kubebuilder:validation:Optional
	PrometheusReport PrometheusReport `json:"prometheus"`
	// +kubebuilder:validation:Optional
	GrafanaReport GrafanaReport `json:"grafana"`
}

type PrometheusMetricType string

const (
	PrometheusMetricTypeCounter PrometheusMetricType = "counter"
	PrometheusMetricTypeRate    PrometheusMetricType = "rate"
)

type PrometheusMetric struct {
	Name string               `json:"name"`
	Help string               `json:"help"`
	Type PrometheusMetricType `json:"type"`
}

type PrometheusReport struct {
	Address string `json:"address"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=10
	DumpFrequencySeconds int `json:"dumpFrequency"`
	// +kubebuilder:validation:Optional
	InstanceUp PrometheusMetric `json:"instanceUp"`
	// +kubebuilder:validation:Optional
	InstanceUpRate PrometheusMetric `json:"instanceUpRate"`
}

func (p *PrometheusReport) Get(name string, rate int) PrometheusReport {
	report := p.DeepCopy()
	if reflect.TypeOf(report.InstanceUp) == reflect.TypeOf(PrometheusMetric{}) {
		report.InstanceUp = *DefaultTotalRunningInstanceMetric(name, rate)
	}
	if reflect.TypeOf(report.InstanceUpRate) == reflect.TypeOf(PrometheusMetric{}) {
		report.InstanceUpRate = *DefaultTotalRunningInstanceRateMetric(name, rate)
	}
	return *report
}

func DefaultTotalRunningInstanceMetric(name string, rate int) *PrometheusMetric {
	return &PrometheusMetric{
		Type: PrometheusMetricTypeCounter,
		Name: fmt.Sprintf(
			"%s_total_running_instance_seconds",
			strings.ReplaceAll(name, "-", "_"),
		),
		Help: fmt.Sprintf(
			"Total number of running instance for %s report recorded every %d seconds",
			name,
			rate,
		),
	}
}

func DefaultTotalRunningInstanceRateMetric(name string, rate int) *PrometheusMetric {
	return &PrometheusMetric{
		Type: PrometheusMetricTypeRate,
		Name: fmt.Sprintf("%s_rate_running_instance_seconds", strings.ReplaceAll(name, "-", "_")),
		Help: fmt.Sprintf(
			"The running instance rate for %s report recorded every %d seconds",
			name,
			rate,
		),
	}
}

type GrafanaReport struct {
}
