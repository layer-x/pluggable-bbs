package middleware

import (
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/runtime-schema/metric"
	"github.com/pivotal-golang/lager"
)

const (
	requestLatency = metric.Duration("RequestLatency")
	requestCount   = metric.Counter("RequestCount")
)

func LogWrap(logger lager.Logger, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestLog := logger.Session("request", lager.Data{
			"method":  r.Method,
			"request": r.URL.String(),
		})

		requestLog.Info("serving")
		handler.ServeHTTP(w, r)
		requestLog.Info("done")
	}
}

func EmitLatency(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		f(w, r)
		requestLatency.Send(time.Since(startTime))
	}
}

func RequestCountWrap(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestCount.Increment()
		handler.ServeHTTP(w, r)
	}
}
