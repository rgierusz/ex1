package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"time"
)

const metricPrefix = "rgierusz_ex1_"

var (
	HealthCheckStartedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: metricPrefix + "healthCheck_started_total",
		Help: "The total number of health check events",
	})

	HealthCheckCompletedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: metricPrefix + "healthCheck_completed_total",
		Help: "The total number of health check events",
	})
)

var (
	GetBalanceStartedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: metricPrefix + "getBalance_started_total",
		Help: "The total number of invocations of get balance handler",
	})

	GetBalanceCompletedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: metricPrefix + "getBalance_completed_total",
		Help: "The total number of completions of get balance invocations handler",
	})
)

var (
	HTTPResponseTimeHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    metricPrefix + "http_request_sec_duration",
		Help:    "Histogram of response time for HTTP requests in milliseconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"path", "method"})
)

var (
	ETHCallCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricPrefix + "eth_call_total",
			Help: "The total number ETH calls",
		},
		[]string{"error"},
	)
)

var (
	GenericResponseErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricPrefix + "response_error_total",
			Help: "The total numbers of response errors",
		},
		[]string{"error"},
	)
)

func HandlerMetricsWrapper(sc prometheus.Counter, cc prometheus.Counter, h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sc.Inc()

		h(w, r)

		cc.Inc()
		HTTPResponseTimeHistogram.WithLabelValues(r.RequestURI, r.Method).Observe(time.Since(start).Seconds())
	}
}
