package cadvisor

import (
	"context"
	"fmt"
	"github.com/google/cadvisor/cache/memory"
	_ "github.com/google/cadvisor/container/containerd/install"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/utils/sysfs"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	"net/http"
	"time"
)

const (
	statsCacheDuration           = 2 * time.Minute
	maxHousekeepingInterval      = 15 * time.Second
	allowDynamicHousekeeping     = true
	KubernetesPodNameLabel       = "io.kubernetes.pod.name"
	KubernetesPodNamespaceLabel  = "io.kubernetes.pod.namespace"
	KubernetesContainerNameLabel = "io.kubernetes.container.name"
)

type Server struct {
	manager.Manager
	collector   *Collector
	srv         *http.Server
	Port        int
	MetricsPath string
}

func New(metricsPath string, port int, cgroupRoot []string) (*Server, error) {
	sysFs := sysfs.NewRealSysFs()

	duration := maxHousekeepingInterval
	housekeepingConfig := manager.HousekeepingConfig{
		Interval:     &duration,
		AllowDynamic: pointer.BoolPtr(allowDynamicHousekeeping),
	}

	m, err := manager.New(memory.New(statsCacheDuration, nil), sysFs, housekeepingConfig, defaultIncludedMetrics, http.DefaultClient, cgroupRoot, nil /* containerEnvMetadataWhiteList */, "" /* perfEventsFile */, time.Duration(0) /*resctrlInterval*/)
	if err != nil {
		return nil, err
	}

	c := NewCollector(m, containerLabels)

	return &Server{
		Manager:     m,
		collector:   &c,
		MetricsPath: metricsPath,
		Port:        port,
	}, nil
}

func (s *Server) Run() {

	promRegistry := prometheus.NewRegistry()
	promRegistry.Register(s.collector)

	http.Handle(s.MetricsPath, promhttp.InstrumentMetricHandler(
		promRegistry, promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{EnableOpenMetrics: false}),
	))

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: http.DefaultServeMux,
	}

	go func() {

		if err := s.Manager.Start(); err != nil {
			klog.Fatalf("Start cadvisor manager failed: %v", err)
		}

		if err := s.srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				klog.Fatalf("Start http server failed: %v", err)
			}
		}
		klog.Info("exit http server")
	}()

	return
}

func (s *Server) Shutdown() {

	if err := s.Manager.Stop(); err != nil {
		klog.Fatalf("Start cadvisor manager failed: %v", err)
	}

	if s.srv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		klog.ErrorS(err, "Close old http server failed, skip")
	}

	return
}
