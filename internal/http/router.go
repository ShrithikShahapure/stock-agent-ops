package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/handlers"
	"github.com/shrithkshahapure/stock-agent-ops/internal/metrics"
	"github.com/shrithkshahapure/stock-agent-ops/internal/middleware"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/cache"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
	redisclient "github.com/shrithkshahapure/stock-agent-ops/internal/services/redis"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// Server holds dependencies for the HTTP server
type Server struct {
	cfg         *config.Config
	redis       *redisclient.Client
	metrics     *metrics.Metrics
	registry    *prometheus.Registry
	router      *chi.Mux
	runner      *python.Runner
	taskManager *tasks.Manager
	cache       *cache.Cache
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, redis *redisclient.Client, m *metrics.Metrics) *Server {
	// Use passed metrics or create new ones
	metricsInstance := m
	var registry *prometheus.Registry
	if metricsInstance == nil {
		registry = prometheus.NewRegistry()
		metricsInstance = metrics.New(registry)
	} else {
		registry = metricsInstance.Registry()
	}

	// Create Python runner
	runner := python.NewRunner(cfg)

	// Create task manager
	taskManager := tasks.NewManager(cfg, runner, redis, metricsInstance)

	// Create cache service (24 hour TTL)
	cacheService := cache.NewCache(redis, metricsInstance, 24*time.Hour)

	s := &Server{
		cfg:         cfg,
		redis:       redis,
		metrics:     metricsInstance,
		registry:    registry,
		router:      chi.NewRouter(),
		runner:      runner,
		taskManager: taskManager,
		cache:       cacheService,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures middleware stack
func (s *Server) setupMiddleware() {
	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Recovery from panics
	s.router.Use(middleware.Recovery)

	// Request logging
	s.router.Use(middleware.Logger)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Create handlers
	healthHandler := handlers.NewHealthHandler(s.cfg)
	trainHandler := handlers.NewTrainHandler(s.cfg, s.taskManager)
	predictHandler := handlers.NewPredictHandler(s.cfg, s.runner, s.cache, s.taskManager, s.metrics)
	analyzeHandler := handlers.NewAnalyzeHandler(s.runner)
	statusHandler := handlers.NewStatusHandler(s.cfg, s.taskManager)
	monitorHandler := handlers.NewMonitorHandler(s.cfg, s.runner)
	systemHandler := handlers.NewSystemHandler(s.cfg, s.redis, s.cache)
	outputsHandler := handlers.NewOutputsHandler(s.cfg)

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(s.redis)

	// Health and info endpoints
	s.router.Get("/", healthHandler.Root)
	s.router.Get("/health", healthHandler.Health)
	s.router.Get("/docs", healthHandler.Docs)
	s.router.Get("/openapi.json", healthHandler.OpenAPI)

	// Prometheus metrics
	s.router.Handle("/metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))

	// Training endpoints (5/hour rate limit)
	s.router.With(rateLimiter.Limit(5, time.Hour, "train_parent")).Post("/train-parent", trainHandler.TrainParent)
	s.router.With(rateLimiter.Limit(5, time.Hour, "train_child")).Post("/train-child", trainHandler.TrainChild)

	// Prediction endpoints (40/hour rate limit)
	s.router.With(rateLimiter.Limit(40, time.Hour, "predict_parent")).Post("/predict-parent", predictHandler.PredictParent)
	s.router.With(rateLimiter.Limit(40, time.Hour, "predict_child")).Post("/predict-child", predictHandler.PredictChild)

	// Analysis
	s.router.Post("/analyze", analyzeHandler.Analyze)

	// Status
	s.router.Get("/status/{task_id}", statusHandler.GetStatus)

	// Monitoring
	s.router.Post("/monitor/parent", monitorHandler.MonitorParent)
	s.router.Post("/monitor/{ticker}", monitorHandler.MonitorTicker)
	s.router.Get("/monitor/{ticker}/drift", monitorHandler.GetDrift)
	s.router.Get("/monitor/{ticker}/eval", monitorHandler.GetEval)

	// System
	s.router.Get("/system/logs", systemHandler.GetLogs)
	s.router.Get("/system/cache", systemHandler.GetCache)
	s.router.Delete("/system/reset", systemHandler.Reset)

	// Outputs
	s.router.Get("/outputs", outputsHandler.ListOutputs)
	s.router.Get("/outputs/{ticker}", outputsHandler.ListTickerOutputs)
}

// Router returns the chi router
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Metrics returns the metrics instance
func (s *Server) Metrics() *metrics.Metrics {
	return s.metrics
}
