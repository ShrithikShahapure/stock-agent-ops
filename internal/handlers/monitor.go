package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
)

// MonitorHandler handles monitoring endpoints
type MonitorHandler struct {
	cfg    *config.Config
	runner *python.Runner
}

// NewMonitorHandler creates a new monitor handler
func NewMonitorHandler(cfg *config.Config, runner *python.Runner) *MonitorHandler {
	return &MonitorHandler{
		cfg:    cfg,
		runner: runner,
	}
}

// MonitorParent handles POST /monitor/parent
func (h *MonitorHandler) MonitorParent(w http.ResponseWriter, r *http.Request) {
	result, err := h.runner.MonitorParent(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Monitoring failed: " + err.Error(),
		})
		return
	}

	// Add links
	if result.Data != nil {
		ticker := h.cfg.ParentTicker
		result.Data["links"] = map[string]string{
			"get_drift_json": "/monitor/" + ticker + "/drift",
			"get_eval_json":  "/monitor/" + ticker + "/eval",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Data)
}

// MonitorTicker handles POST /monitor/{ticker}
func (h *MonitorHandler) MonitorTicker(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	ticker = strings.TrimSpace(strings.ToUpper(ticker))

	result, err := h.runner.MonitorTicker(r.Context(), ticker)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Monitoring failed: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Data)
}

// GetDrift handles GET /monitor/{ticker}/drift
func (h *MonitorHandler) GetDrift(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	ticker = strings.ToLower(ticker)

	driftDir := filepath.Join(h.cfg.OutputsDir, ticker, "drift")
	if _, err := os.Stat(driftDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "No drift report found",
		})
		return
	}

	jsonPath := filepath.Join(driftDir, "latest_drift.json")
	if _, err := os.Stat(jsonPath); err == nil {
		// Read and return JSON file
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"detail": "Failed to read drift report: " + err.Error(),
			})
			return
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"detail": "Invalid drift report format",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// List files in drift directory
	files, err := os.ReadDir(driftDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Failed to read drift directory",
		})
		return
	}

	fileNames := make([]string, 0, len(files))
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files":   fileNames,
		"message": "Access HTML report in outputs/",
		"detail":  "JSON summary missing.",
	})
}

// GetEval handles GET /monitor/{ticker}/eval
func (h *MonitorHandler) GetEval(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	ticker = strings.ToLower(ticker)

	evalPath := filepath.Join(h.cfg.OutputsDir, ticker, "agent_eval", "latest_eval.json")
	if _, err := os.Stat(evalPath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "No evaluation found. Run POST /monitor/{ticker} first.",
		})
		return
	}

	data, err := os.ReadFile(evalPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Failed to read evaluation: " + err.Error(),
		})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Invalid evaluation format",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
