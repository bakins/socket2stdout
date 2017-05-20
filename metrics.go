package socket

import "github.com/prometheus/client_golang/prometheus"

const metricsNamespace = "socket2stdout"

var (
	connectionGuage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "connections_current",
			Help:      "current connections",
		},
	)

	connectionCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "connections_total",
			Help:      "total connections",
		},
	)

	linesRead = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "lines_read",
			Help:      "total lines read",
		},
	)

	linesWritten = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "lines_written",
			Help:      "total lines written",
		},
	)
)

func init() {
	prometheus.MustRegister(connectionGuage)
	prometheus.MustRegister(connectionCounter)
	prometheus.MustRegister(linesRead)
	prometheus.MustRegister(linesWritten)
}
