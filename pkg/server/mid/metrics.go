package mid

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		},
		[]string{"method", "endpoint", "handler", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "handler", "status_code"},
	)
)

func MetricsMiddleware(metricsURL string, engine *gin.Engine) gin.HandlerFunc {
	engine.GET(metricsURL, gin.WrapH(promhttp.Handler()))
	return func(context *gin.Context) {
		if shouldSkipURL(context.Request.URL.Path) {
			context.Next()
			return
		}
		startTime := time.Now()
		context.Next()
		duration := time.Since(startTime)
		handlerName := context.HandlerName()
		endpoint := context.FullPath()
		method := context.Request.Method
		statusCode := strconv.Itoa(context.Writer.Status())
		if endpoint != "" && statusCode != "404" {
			httpRequestsTotal.WithLabelValues(method, endpoint, handlerName, statusCode).Inc()
			httpRequestDuration.WithLabelValues(method, endpoint, handlerName, statusCode).Observe(duration.Seconds())
		}
	}
}
