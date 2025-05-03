package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"github.com/Lesion45/go-load-balancer/internal/config"
	backendsCfg "github.com/Lesion45/go-load-balancer/internal/config/backends"
	"github.com/Lesion45/go-load-balancer/internal/healthcheck"
	"github.com/Lesion45/go-load-balancer/internal/loadbalancer"
	"github.com/Lesion45/go-load-balancer/internal/loadbalancer/strategies"
	p "github.com/Lesion45/go-load-balancer/internal/prometheus"
	"github.com/Lesion45/go-load-balancer/internal/utils"
	"github.com/Lesion45/go-load-balancer/pkg/logger"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type App struct {
	Log         *zap.Logger
	Backends    []*backend.SimpleBackend
	Balancer    strategies.BalancerStrategy
	HealthCheck healthcheck.HealthChecker
	Server      *http.Server
}

func NewApp() *App {
	cfg := config.MustLoad()
	log := logger.NewZap(cfg.Env)

	// Init prometheus
	p.MustLoad()

	// Load backends config
	backendsConfig := backendsCfg.MustLoad()

	// Init backends
	var backends []*backend.SimpleBackend
	for _, backendConfig := range backendsConfig.Backends {
		b, err := backend.NewSimpleBackend(
			backendConfig.Address,
			cfg.RateLimitingConfig.StrategyName,
			backendConfig.Burst,
			backendConfig.Rate,
		)
		if err != nil {
			log.Error("Failed to create backend",
				zap.String("name", backendConfig.Name),
				zap.Error(err))
			continue
		}
		backends = append(backends, b)
	}

	// Init balancer
	settings := &loadbalancer.Settings{
		StrategyName: cfg.LoadBalancerConfig.StrategyName,
		Backends:     backends,
	}
	balancer := strategies.NewBalancerStrategy(log, settings)

	// Init health checker
	healthChecker := healthcheck.NewHTTPHealthChecker(log,
		&http.Client{},
		cfg.HealthCheckConfig.Endpoint,
		cfg.HealthCheckConfig.BaseInterval,
		cfg.HealthCheckConfig.MaxInterval)

	app := &App{
		Log:         log,
		Backends:    backends,
		Balancer:    balancer,
		HealthCheck: healthChecker,
	}

	// Init server
	app.Server = &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: app.createRouter(),
	}

	return app
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		a.Log.Info("Starting server", zap.String("addr", a.Server.Addr))
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	time.Sleep(10 * time.Millisecond)
	a.Log.Info("Starting health checker")
	a.HealthCheck.Start(ctx, a.Backends)

	<-quit
	a.Log.Info("Shutdown signal received")
	cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Log.Fatal("Server shutdown failed", zap.Error(err))
	}
	a.Log.Info("Server exited properly")
}

func (a *App) createRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", a.handleRequest)
	return mux
}

func (a *App) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := uuid.NewString()

	a.Log.Info("Incoming request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote", r.RemoteAddr),
		zap.String("request_id", requestID),
	)

	backend, err := a.Balancer.NextBackend()
	if err != nil {
		a.Log.Error("No available backend",
			zap.Error(err),
			zap.String("request_id", requestID),
		)

		p.Observe(r.Method, "none", "503", start)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	if !backend.Allow() {
		a.Log.Warn("Rate limit exceeded",
			zap.String("backend", backend.GetUrl()),
			zap.String("request_id", requestID),
		)

		p.Observe(r.Method, backend.GetUrl(), "429", start)
		http.Error(w, "Too many requests, please try again later", http.StatusTooManyRequests)
		return
	}

	URL, err := url.Parse(backend.GetUrl())
	if err != nil {
		a.Log.Error("Failed to parse backend URL",
			zap.String("backend", backend.GetUrl()),
			zap.Error(err),
			zap.String("request_id", requestID),
		)

		p.Observe(r.Method, backend.GetUrl(), "502", start)
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}

	a.Log.Info("Proxying request",
		zap.String("backend", URL.String()),
		zap.String("request_id", requestID),
	)

	proxy := httputil.NewSingleHostReverseProxy(URL)

	rec := &utils.StatusRecorder{ResponseWriter: w, Status: 200}
	proxy.ServeHTTP(rec, r)

	duration := time.Since(start)
	statusCode := rec.Status

	p.Observe(r.Method, backend.GetUrl(), fmt.Sprintf("%d", rec.Status), start)

	a.Log.Info("Request completed",
		zap.Int("status", statusCode),
		zap.Duration("duration", duration),
		zap.String("backend", backend.GetUrl()),
		zap.String("request_id", requestID),
	)
}
