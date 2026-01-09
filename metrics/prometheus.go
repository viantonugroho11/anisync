package metrics

import "github.com/prometheus/client_golang/prometheus"

var acquireCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "anisync",
		Name:      "lock_acquire_total",
	},
	[]string{"success"},
)

var locksHeld = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "anisync",
		Name:      "locks_held",
	},
)

func init() {
	prometheus.MustRegister(acquireCounter)
	prometheus.MustRegister(locksHeld)
}

func RecordAcquire(success bool) {
	acquireCounter.WithLabelValues(
		func() string {
			if success {
				return "true"
			}
			return "false"
		}(),
	).Inc()
	if success {
		locksHeld.Inc()
	}
}

func RecordRelease(success bool) {
	if success {
		locksHeld.Dec()
	}
}
