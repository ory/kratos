package prometheus

import "github.com/prometheus/client_golang/prometheus"

// Metrics prototypes
type Metrics struct {
	ResponseTime *prometheus.HistogramVec
}

// Method for creation new custom Prometheus  metrics
func NewMetrics(version, hash, date string) *Metrics {
	pm := &Metrics{
		ResponseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "kratos_response_time_seconds",
				Help: "Description",
				ConstLabels: map[string]string{
					"version":   version,
					"hash":      hash,
					"buildTime": date,
				},
			},
			[]string{"endpoint"},
		),
	}
	err := prometheus.Register(pm.ResponseTime)

	if err != nil {
		panic(err)
	}
	return pm
}
