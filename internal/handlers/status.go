package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// StatusHandler handles status endpoints
type StatusHandler struct {
	cfg         *config.Config
	taskManager tasks.ManagerInterface
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(cfg *config.Config, taskManager tasks.ManagerInterface) *StatusHandler {
	return &StatusHandler{
		cfg:         cfg,
		taskManager: taskManager,
	}
}

// modelExists checks if a model file exists
func (h *StatusHandler) modelExists(ticker string, modelType string) bool {
	var path string
	if modelType == "parent" {
		path = filepath.Join(h.cfg.ParentDir, h.cfg.ParentTicker+"_parent_model.pt")
	} else {
		path = filepath.Join(h.cfg.OutputsDir, strings.ToUpper(ticker), strings.ToUpper(ticker)+"_child_model.pt")
	}
	_, err := os.Stat(path)
	return err == nil
}

// GetStatus handles GET /status/{task_id}
func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")
	taskID = strings.ToLower(taskID)

	// Map "parent" to "parent_training"
	if taskID == "parent" {
		taskID = "parent_training"
	}

	// Determine model type for disk check
	tickerForDisk := taskID
	modelType := "child"
	if taskID == "parent_training" {
		tickerForDisk = "parent"
		modelType = "parent"
	}
	fileExists := h.modelExists(tickerForDisk, modelType)

	// Get status from task manager
	var status *tasks.TaskStatus
	if h.taskManager != nil {
		status = h.taskManager.GetStatus(taskID)
	}

	// If no status in Redis but file exists, return completed
	if status == nil {
		if fileExists {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "completed",
				"detail":  "Model file found on disk",
				"task_id": taskID,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Task '" + taskID + "' not found.",
		})
		return
	}

	// If status is not running/failed but file exists, mark as completed
	if status.Status != "running" && status.Status != "failed" && fileExists {
		status.Status = "completed"
	}

	// Build response
	response := map[string]interface{}{
		"status":  status.Status,
		"task_id": taskID,
	}

	if status.Result != nil {
		response["result"] = status.Result
	}
	if status.Error != "" {
		response["error"] = status.Error
	}
	if status.CompletedAt != "" {
		response["completed_at"] = status.CompletedAt
	}
	if status.FailedAt != "" {
		response["failed_at"] = status.FailedAt
	}

	// Calculate elapsed seconds for running tasks
	if status.Status == "running" && status.StartTime != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", status.StartTime)
		if err == nil {
			response["elapsed_seconds"] = int(time.Since(startTime).Seconds())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
