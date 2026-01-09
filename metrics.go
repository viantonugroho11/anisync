package anisync

import "github.com/prometheus/client_golang/prometheus"

var acquireCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "anisync",
		Name:      "lock_acquire_total",
	},
	[]string{"success"},
)

func init() {
	prometheus.MustRegister(acquireCounter)
}

func observeAcquire(success bool) {
	acquireCounter.WithLabelValues(
		func() string {
			if success {
				return "true"
			}
			return "false"
		}(),
	).Inc()
}
