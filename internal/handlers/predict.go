package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/metrics"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/cache"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// PredictHandler handles prediction endpoints
type PredictHandler struct {
	cfg         *config.Config
	runner      python.RunnerInterface
	cache       cache.CacheInterface
	taskManager tasks.ManagerInterface
	metrics     *metrics.Metrics
}

// NewPredictHandler creates a new predict handler
func NewPredictHandler(cfg *config.Config, runner python.RunnerInterface, c cache.CacheInterface, taskManager tasks.ManagerInterface, m *metrics.Metrics) *PredictHandler {
	return &PredictHandler{
		cfg:         cfg,
		runner:      runner,
		cache:       c,
		taskManager: taskManager,
		metrics:     m,
	}
}

// modelExists checks if a model file exists
func (h *PredictHandler) modelExists(ticker string, modelType string) bool {
	var path string
	if modelType == "parent" {
		path = filepath.Join(h.cfg.ParentDir, h.cfg.ParentTicker+"_parent_model.pt")
	} else {
		path = filepath.Join(h.cfg.OutputsDir, strings.ToUpper(ticker), strings.ToUpper(ticker)+"_child_model.pt")
	}
	_, err := os.Stat(path)
	return err == nil
}

// PredictParent handles POST /predict-parent
func (h *PredictHandler) PredictParent(w http.ResponseWriter, r *http.Request) {
	if h.metrics != nil {
		h.metrics.PredictionTotal.WithLabelValues("parent").Inc()
	}
	start := time.Now()

	result, err := h.runner.PredictParent(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": err.Error(),
		})
		return
	}

	if h.metrics != nil {
		h.metrics.PredictionLatency.WithLabelValues("parent").Observe(time.Since(start).Seconds())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result.Data,
	})
}

// PredictChild handles POST /predict-child
func (h *PredictHandler) PredictChild(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		Ticker string `json:"ticker"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Invalid request body",
		})
		return
	}

	ticker := strings.TrimSpace(strings.ToUpper(req.Ticker))
	if ticker == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "ticker is required",
		})
		return
	}

	taskID := strings.ToLower(ticker)

	if h.metrics != nil {
		h.metrics.PredictionTotal.WithLabelValues("child").Inc()
	}
	start := time.Now()

	// Check cache first
	if h.cache != nil {
		if cached, found := h.cache.Get(ticker); found {
			if h.metrics != nil {
				h.metrics.PredictionLatency.WithLabelValues("child").Observe(time.Since(start).Seconds())
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": cached,
			})
			return
		}
	}

	// Execute prediction
	result, err := h.runner.PredictChild(r.Context(), ticker)
	if err != nil {
		errStr := err.Error()

		// Check if model missing - trigger auto-training
		if strings.Contains(strings.ToLower(errStr), "missing") || strings.Contains(strings.ToLower(errStr), "not found") {
			// Check parent model
			if !h.modelExists("parent", "parent") {
				if h.taskManager != nil {
					parentStatus := h.taskManager.GetStatus("parent_training")
					if parentStatus == nil || parentStatus.Status != "completed" {
						h.taskManager.StartTrainParent()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
						json.NewEncoder(w).Encode(map[string]interface{}{
							"status":  "training",
							"detail":  "Parent model missing. Training parent first.",
							"task_id": "parent_training",
						})
						return
					}
				}
			}

			// Check if child training is already running
			if h.taskManager != nil && h.taskManager.IsRunning(taskID) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "training",
					"detail":  "Training in progress. Please retry later.",
					"task_id": taskID,
				})
				return
			}

			// Start child training with cache refresh chain
			if h.taskManager != nil {
				chainFn := func() {
					// After training, cache the prediction
					ctx := context.Background()
					res, err := h.runner.PredictChild(ctx, ticker)
					if err == nil && h.cache != nil {
						h.cache.Set(ticker, res.Data)
					}
				}
				h.taskManager.StartTrainChild(taskID, chainFn)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "training",
				"detail":  "Model for " + ticker + " missing. Training started (with auto-prediction).",
				"task_id": taskID,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": errStr,
		})
		return
	}

	// Cache result
	if h.cache != nil {
		h.cache.Set(ticker, result.Data)
	}

	if h.metrics != nil {
		h.metrics.PredictionLatency.WithLabelValues("child").Observe(time.Since(start).Seconds())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result.Data,
	})
}
