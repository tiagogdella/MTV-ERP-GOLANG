package metrics

import (
	"net/http"
	"strconv"
	"time"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total de requisições HTTP recebidas pelo gateway.",
	}, []string{"method", "route", "status"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duração das requisições HTTP, em segundos.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "desconhecida"
		}

		if len(route) > 1 {
			route = strings.TrimSuffix(route, "/")
		}

		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}

		requestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(status)).Inc()
		requestDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}
