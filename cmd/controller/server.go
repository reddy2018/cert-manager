package main

import (
	"context"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusMetricsServerAddress         = "0.0.0.0:9402"
	prometheusMetricsServerShutdownTimeout = 10 * time.Second
	prometheusMetricsServerReadTimeout     = 8 * time.Second
	prometheusMetricsServerWriteTimeout    = 8 * time.Second
	prometheusMetricsServerMaxHeaderBytes  = 1 << 20
)

type prometheusMetricsServer struct {
	http.Server
}

func prometheusHandler() http.Handler {
	return prometheus.Handler()
}

func newPrometheusMetricsServer() *prometheusMetricsServer {

	r := mux.NewRouter()
	r.Handle("/metrics", prometheusHandler())

	// Create server and register prometheus metrics handler
	s := &prometheusMetricsServer{
		Server: http.Server{
			Addr:           prometheusMetricsServerAddress,
			ReadTimeout:    prometheusMetricsServerReadTimeout,
			WriteTimeout:   prometheusMetricsServerWriteTimeout,
			MaxHeaderBytes: prometheusMetricsServerMaxHeaderBytes,
			Handler:        r,
		},
	}

	return s
}

func (s *prometheusMetricsServer) WaitShutdown(stopCh <-chan struct{}) {
	<-stopCh
	glog.Info("Stopping Prometheus metrics server...")

	ctx, cancel := context.WithTimeout(context.Background(), prometheusMetricsServerShutdownTimeout)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		glog.Errorf("Prometheus metrics server shutdown error: %v", err)
		return
	}

	glog.Info("Prometheus metrics server gracefully stopped")
}

func StartPrometheusMetricsServer(stopCh <-chan struct{}) {
	s := newPrometheusMetricsServer()

	go func() {

		glog.Infof("Listening on http://%s", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			glog.Errorf("Error running prometheus metrics server: %s", err.Error())
			return
		}

		glog.Infof("Prometheus metrics server exited")

	}()

	s.WaitShutdown(stopCh)
}
