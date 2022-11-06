package messages

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var histogramResponseTime = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "route256",
		Subsystem: "telegram",
		Name:      "histogram_response_time_seconds",
		Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
	},
	[]string{"status"},
)

func observeResponse(elapsed time.Duration, err bool) {
	histogramResponseTime.
		WithLabelValues(strconv.FormatBool(err)).
		Observe(elapsed.Seconds())
}
