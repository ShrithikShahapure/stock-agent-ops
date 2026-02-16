package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
)

// AnalyzeHandler handles the analyze endpoint
type AnalyzeHandler struct {
	runner *python.Runner
}

// NewAnalyzeHandler creates a new analyze handler
func NewAnalyzeHandler(runner *python.Runner) *AnalyzeHandler {
	return &AnalyzeHandler{runner: runner}
}

// Analyze handles POST /analyze
func (h *AnalyzeHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		Ticker   string `json:"ticker"`
		UseFMI   bool   `json:"use_fmi"`
		ThreadID string `json:"thread_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Invalid request body",
		})
		return
	}

	ticker := strings.TrimSpace(req.Ticker)
	if ticker == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Ticker required",
		})
		return
	}

	result, err := h.runner.Analyze(r.Context(), ticker, req.ThreadID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Analysis failed: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Data)
}
