package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// TrainHandler handles training endpoints
type TrainHandler struct {
	cfg         *config.Config
	taskManager *tasks.Manager
}

// NewTrainHandler creates a new train handler
func NewTrainHandler(cfg *config.Config, taskManager *tasks.Manager) *TrainHandler {
	return &TrainHandler{
		cfg:         cfg,
		taskManager: taskManager,
	}
}

// modelExists checks if a model file exists
func (h *TrainHandler) modelExists(ticker string, modelType string) bool {
	var path string
	if modelType == "parent" {
		path = filepath.Join(h.cfg.ParentDir, h.cfg.ParentTicker+"_parent_model.pt")
	} else {
		path = filepath.Join(h.cfg.OutputsDir, strings.ToUpper(ticker), strings.ToUpper(ticker)+"_child_model.pt")
	}
	_, err := os.Stat(path)
	return err == nil
}

// TrainParent handles POST /train-parent
func (h *TrainHandler) TrainParent(w http.ResponseWriter, r *http.Request) {
	taskID := "parent_training"

	// Check if model already exists
	if h.modelExists("parent", "parent") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "completed",
			"task_id": taskID,
			"detail":  "Parent model already exists",
		})
		return
	}

	// Check if already running
	if h.taskManager != nil && h.taskManager.IsRunning(taskID) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "already running",
			"task_id": taskID,
		})
		return
	}

	// Start training
	if h.taskManager != nil {
		started, _ := h.taskManager.StartTrainParent()
		if !started {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "already running",
				"task_id": taskID,
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"task_id": taskID,
	})
}

// TrainChild handles POST /train-child
func (h *TrainHandler) TrainChild(w http.ResponseWriter, r *http.Request) {
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

	// Check if parent model exists
	if !h.modelExists("parent", "parent") {
		// Check if parent is already training
		if h.taskManager != nil {
			parentStatus := h.taskManager.GetStatus("parent_training")
			if parentStatus == nil || parentStatus.Status != "completed" {
				h.taskManager.StartTrainParent()
				if h.taskManager.IsRunning("parent_training") {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "started_parent",
						"task_id": "parent_training",
						"detail":  "Parent model missing. Training parent first.",
					})
					return
				}
			}
		}
	}

	// Check if child model already exists
	if h.modelExists(ticker, "child") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "completed",
			"task_id": taskID,
			"detail":  "Model already exists",
		})
		return
	}

	// Check if already running
	if h.taskManager != nil && h.taskManager.IsRunning(taskID) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "running",
			"task_id": taskID,
			"detail":  "Training already in progress",
		})
		return
	}

	// Start training (with chain prediction)
	if h.taskManager != nil {
		h.taskManager.StartTrainChild(taskID, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"task_id": taskID,
	})
}
