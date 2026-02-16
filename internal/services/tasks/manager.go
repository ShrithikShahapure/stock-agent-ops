package tasks

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/metrics"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
	redisclient "github.com/shrithkshahapure/stock-agent-ops/internal/services/redis"
)

// TaskStatus represents the status of a background task
type TaskStatus struct {
	Status      string                 `json:"status"` // running, completed, failed
	StartTime   string                 `json:"start_time,omitempty"`
	CompletedAt string                 `json:"completed_at,omitempty"`
	FailedAt    string                 `json:"failed_at,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// Manager manages background training tasks
type Manager struct {
	runner     *python.Runner
	redis      *redisclient.Client
	metrics    *metrics.Metrics
	maxWorkers int
	sem        chan struct{}
	mu         sync.Mutex
}

// NewManager creates a new task manager
func NewManager(cfg *config.Config, runner *python.Runner, redis *redisclient.Client, m *metrics.Metrics) *Manager {
	return &Manager{
		runner:     runner,
		redis:      redis,
		metrics:    m,
		maxWorkers: cfg.MaxWorkers,
		sem:        make(chan struct{}, cfg.MaxWorkers),
	}
}

// GetStatus retrieves the status of a task from Redis
func (m *Manager) GetStatus(taskID string) *TaskStatus {
	if m.redis == nil {
		return nil
	}

	ctx := context.Background()
	key := redisclient.TaskKey(taskID)

	val, err := m.redis.Get(ctx, key)
	if err != nil {
		return nil
	}

	var status TaskStatus
	if err := json.Unmarshal([]byte(val), &status); err != nil {
		return nil
	}

	return &status
}

// saveStatus saves task status to Redis
func (m *Manager) saveStatus(taskID string, status TaskStatus, ttl time.Duration) {
	if m.redis == nil {
		return
	}

	ctx := context.Background()
	key := redisclient.TaskKey(taskID)

	data, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal task status: %v", err)
		return
	}

	if err := m.redis.Set(ctx, key, string(data), ttl); err != nil {
		log.Printf("Failed to save task status: %v", err)
	}
}

// IsRunning checks if a task is currently running
func (m *Manager) IsRunning(taskID string) bool {
	status := m.GetStatus(taskID)
	return status != nil && status.Status == "running"
}

// StartTrainParent starts parent model training in the background
func (m *Manager) StartTrainParent() (bool, error) {
	taskID := "parent_training"

	// Check if already running
	if m.IsRunning(taskID) {
		return false, nil
	}

	// Acquire semaphore slot
	select {
	case m.sem <- struct{}{}:
		// Got a slot
	default:
		return false, nil // All workers busy
	}

	// Set running status
	m.saveStatus(taskID, TaskStatus{
		Status:    "running",
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
	}, 2*time.Hour)

	if m.metrics != nil {
		m.metrics.TrainingStatus.WithLabelValues(taskID).Set(1)
	}

	// Run in background
	go func() {
		defer func() { <-m.sem }()
		m.runTrainParent(taskID)
	}()

	return true, nil
}

func (m *Manager) runTrainParent(taskID string) {
	start := time.Now()
	ctx := context.Background()

	result, err := m.runner.TrainParent(ctx)

	if err != nil {
		m.saveStatus(taskID, TaskStatus{
			Status:   "failed",
			Error:    err.Error(),
			FailedAt: time.Now().Format("2006-01-02 15:04:05"),
		}, time.Hour)

		if m.metrics != nil {
			m.metrics.TrainingStatus.WithLabelValues(taskID).Set(0)
		}
		log.Printf("Training task %s failed: %v", taskID, err)
		return
	}

	duration := time.Since(start)
	m.saveStatus(taskID, TaskStatus{
		Status:      "completed",
		Result:      result.Data,
		CompletedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, time.Hour)

	if m.metrics != nil {
		m.metrics.TrainingStatus.WithLabelValues(taskID).Set(2)
		m.metrics.TrainingDuration.WithLabelValues(taskID).Observe(duration.Seconds())

		// Update MSE if available
		if result.Data != nil {
			if mse, ok := result.Data["mse"].(float64); ok {
				m.metrics.TrainingMSE.Set(mse)
			}
		}
	}

	log.Printf("Training task %s completed in %v", taskID, duration)
}

// StartTrainChild starts child model training in the background
func (m *Manager) StartTrainChild(ticker string, chainFn func()) (bool, error) {
	taskID := ticker

	// Check if already running
	if m.IsRunning(taskID) {
		return false, nil
	}

	// Acquire semaphore slot
	select {
	case m.sem <- struct{}{}:
		// Got a slot
	default:
		return false, nil
	}

	// Set running status
	m.saveStatus(taskID, TaskStatus{
		Status:    "running",
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
	}, 2*time.Hour)

	if m.metrics != nil {
		m.metrics.TrainingStatus.WithLabelValues(taskID).Set(1)
	}

	// Run in background
	go func() {
		defer func() { <-m.sem }()
		m.runTrainChild(taskID, ticker, chainFn)
	}()

	return true, nil
}

func (m *Manager) runTrainChild(taskID, ticker string, chainFn func()) {
	start := time.Now()
	ctx := context.Background()

	result, err := m.runner.TrainChild(ctx, ticker)

	if err != nil {
		m.saveStatus(taskID, TaskStatus{
			Status:   "failed",
			Error:    err.Error(),
			FailedAt: time.Now().Format("2006-01-02 15:04:05"),
		}, time.Hour)

		if m.metrics != nil {
			m.metrics.TrainingStatus.WithLabelValues(taskID).Set(0)
		}
		log.Printf("Training task %s failed: %v", taskID, err)
		return
	}

	// Run chain function if provided
	if chainFn != nil {
		log.Printf("Task %s: Running chained function...", taskID)
		chainFn()
	}

	duration := time.Since(start)
	m.saveStatus(taskID, TaskStatus{
		Status:      "completed",
		Result:      result.Data,
		CompletedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, time.Hour)

	if m.metrics != nil {
		m.metrics.TrainingStatus.WithLabelValues(taskID).Set(2)
		m.metrics.TrainingDuration.WithLabelValues(taskID).Observe(duration.Seconds())

		if result.Data != nil {
			if mse, ok := result.Data["mse"].(float64); ok {
				m.metrics.TrainingMSE.Set(mse)
			}
		}
	}

	log.Printf("Training task %s completed in %v", taskID, duration)
}
