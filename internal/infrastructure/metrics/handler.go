package metrics

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Metrics struct {
	Handler gin.HandlerFunc
	Prom    *Prometheus
}

func NewMetrics(handler http.Handler) *Metrics {
	prom := NewPrometheus()
	prom.RegisterMetrics()
	return &Metrics{
		Handler: MetricsMiddleware(),
	}
}

func (s *Metrics) StartServer(ctx context.Context, addr string) {
	metricsRouter := http.NewServeMux()
	metricsRouter.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    addr,
		Handler: metricsRouter,
	}

	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP server for metrics exited with error: %v", err)
		}
	}()
	log.Infof("HTTP server for metrics listening at %v", addr)

	<-ctx.Done()
	log.Info("Shutting down metrics server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Errorf("HTTP server for metrics Shutdown failed: %v", err)
	} else {
		log.Info("Metrics server gracefully stopped")
	}
}
