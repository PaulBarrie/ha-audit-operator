package v1beta1

type HAReport struct {
	PrometheusReport PrometheusReport `json:"prometheus"`
	GrafanaReport    GrafanaReport    `json:"grafana"`
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
	Address        string           `json:"address"`
	InstanceUp     PrometheusMetric `json:"instanceUp"`
	InstanceUpRate PrometheusMetric `json:"instanceUpRate"`
}

type GrafanaReport struct {
}

type ReportType string

const (
	ReportTypePrometheus ReportType = "prometheus"
	ReportTypeGrafana    ReportType = "grafana"
)
