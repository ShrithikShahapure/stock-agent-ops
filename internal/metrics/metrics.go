package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
// Metric names match exactly with the FastAPI implementation in backend/state.py
type Metrics struct {
	registry *prometheus.Registry

	// System metrics
	SystemCPU  prometheus.Gauge
	SystemRAM  prometheus.Gauge
	SystemDisk prometheus.Gauge

	// Redis metrics
	RedisUp   prometheus.Gauge
	RedisKeys prometheus.Gauge

	// Training metrics
	TrainingStatus   *prometheus.GaugeVec
	TrainingMSE      prometheus.Gauge
	TrainingDuration *prometheus.HistogramVec

	// Prediction metrics
	PredictionTotal   *prometheus.CounterVec
	PredictionLatency *prometheus.HistogramVec

	// Cache metrics
	CacheHit  *prometheus.CounterVec
	CacheMiss *prometheus.CounterVec
}

// New creates and registers all Prometheus metrics
func New(reg *prometheus.Registry) *Metrics {
	factory := promauto.With(reg)

	m := &Metrics{
		registry: reg,
		// System metrics
		SystemCPU: factory.NewGauge(prometheus.GaugeOpts{
			Name: "system_cpu_percent",
			Help: "CPU percent",
		}),
		SystemRAM: factory.NewGauge(prometheus.GaugeOpts{
			Name: "system_ram_used_mb",
			Help: "RAM MB",
		}),
		SystemDisk: factory.NewGauge(prometheus.GaugeOpts{
			Name: "system_disk_used_mb",
			Help: "Disk Used MB",
		}),

		// Redis metrics
		RedisUp: factory.NewGauge(prometheus.GaugeOpts{
			Name: "redis_up",
			Help: "Redis up=1/down=0",
		}),
		RedisKeys: factory.NewGauge(prometheus.GaugeOpts{
			Name: "redis_keys_total",
			Help: "Number of keys in Redis",
		}),

		// Training metrics
		TrainingStatus: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "training_status",
			Help: "0=idle 1=running 2=completed",
		}, []string{"task_id"}),
		TrainingMSE: factory.NewGauge(prometheus.GaugeOpts{
			Name: "training_mse_last",
			Help: "Last training MSE",
		}),
		TrainingDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "training_duration_seconds",
			Help:    "Training duration in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9h
		}, []string{"task_id"}),

		// Prediction metrics
		PredictionTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "prediction_total",
			Help: "Total predictions",
		}, []string{"type"}),
		PredictionLatency: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "prediction_latency_seconds",
			Help:    "Prediction latency",
			Buckets: prometheus.DefBuckets,
		}, []string{"type"}),

		// Cache metrics
		CacheHit: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "redis_cache_hit_total",
			Help: "Cache hits",
		}, []string{"key"}),
		CacheMiss: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "redis_cache_miss_total",
			Help: "Cache misses",
		}, []string{"key"}),
	}

	return m
}

// Registry returns the Prometheus registry
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}
