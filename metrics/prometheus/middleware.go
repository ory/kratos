package prometheus

import (
	"net/http"
	"time"
)

type MetricsManager struct {
	prometheusMetrics *Metrics
}

func NewMetricsManager(version, hash, buildTime string) *MetricsManager {
	return &MetricsManager{
		prometheusMetrics: NewMetrics(version, hash, buildTime),
	}
}

// Main middleware method to collect metrics for Prometheus.
func (pmm *MetricsManager) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)

	pmm.prometheusMetrics.ResponseTime.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
}
