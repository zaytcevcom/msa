package internalhttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_server_requests_seconds",
			Help:    "Histogram of response latency (seconds) of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"uri", "method", "status"},
	)

	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"uri", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestCounter)
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		uri := r.URL.Path
		method := r.Method
		status := strconv.Itoa(rw.statusCode)

		requestDuration.WithLabelValues(uri, method, status).Observe(duration)
		requestCounter.WithLabelValues(uri, method, status).Inc()
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
