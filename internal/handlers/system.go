package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/cache"
	redisclient "github.com/shrithkshahapure/stock-agent-ops/internal/services/redis"
)

// SystemHandler handles system endpoints
type SystemHandler struct {
	cfg   *config.Config
	redis *redisclient.Client
	cache *cache.Cache
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(cfg *config.Config, redis *redisclient.Client, cache *cache.Cache) *SystemHandler {
	return &SystemHandler{
		cfg:   cfg,
		redis: redis,
		cache: cache,
	}
}

// GetLogs handles GET /system/logs
func (h *SystemHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// Parse lines parameter
	linesStr := r.URL.Query().Get("lines")
	lines := 100
	if linesStr != "" {
		if n, err := strconv.Atoi(linesStr); err == nil && n > 0 {
			lines = n
		}
	}

	logDir := h.cfg.LogsDir
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"logs": "Log directory not found.",
		})
		return
	}

	// Find log files
	files, err := os.ReadDir(logDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"logs": "Failed to read log directory.",
		})
		return
	}

	var logFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".log") {
			logFiles = append(logFiles, f.Name())
		}
	}

	if len(logFiles) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"logs": "No log files found.",
		})
		return
	}

	// Sort and get latest
	sort.Strings(logFiles)
	latestFile := logFiles[len(logFiles)-1]
	logPath := filepath.Join(logDir, latestFile)

	// Read file
	data, err := os.ReadFile(logPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to read logs: " + err.Error(),
		})
		return
	}

	// Get last N lines
	content := string(data)
	allLines := strings.Split(content, "\n")
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}
	lastLines := strings.Join(allLines[start:], "\n")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"logs":     lastLines,
		"filename": latestFile,
	})
}

// GetCache handles GET /system/cache
func (h *SystemHandler) GetCache(w http.ResponseWriter, r *http.Request) {
	if h.redis == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Redis not connected",
		})
		return
	}

	ticker := r.URL.Query().Get("ticker")

	if ticker == "" {
		// Return list of cached tickers
		tickers, err := h.cache.GetCachedTickers()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cached_tickers": tickers,
			"count":          len(tickers),
		})
		return
	}

	// Return specific ticker's cache
	ticker = strings.TrimSpace(strings.ToUpper(ticker))
	data, err := h.cache.GetForTicker(ticker)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": err.Error(),
		})
		return
	}

	if data == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "No cache found for " + ticker,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Reset handles DELETE /system/reset
func (h *SystemHandler) Reset(w http.ResponseWriter, r *http.Request) {
	results := make(map[string]string)
	ctx := context.Background()

	// 1. Reset Redis
	if h.redis != nil {
		if err := h.redis.FlushAll(ctx); err != nil {
			results["redis"] = "Failed: " + err.Error()
		} else {
			results["redis"] = "Flushed"
		}
	} else {
		results["redis"] = "Skipped (Not connected)"
	}

	// 2. Reset Qdrant (via Python - simplified, just note it)
	// In a full implementation, we'd call a Python script or Qdrant API
	results["qdrant"] = "Note: Run Python reset separately if needed"

	// 3. Reset Feast
	feastDir := h.cfg.FeatureStoreDir
	filesToRemove := []string{
		filepath.Join(feastDir, "data", "registry.db"),
		filepath.Join(feastDir, "data", "features.parquet"),
		filepath.Join(feastDir, "registry.db"),
		filepath.Join(feastDir, "online_store.db"),
	}

	var removedFeast []string
	for _, p := range filesToRemove {
		if _, err := os.Stat(p); err == nil {
			if err := os.Remove(p); err == nil {
				removedFeast = append(removedFeast, filepath.Base(p))
			}
		}
	}

	if len(removedFeast) > 0 {
		results["feast"] = "Removed: " + strings.Join(removedFeast, ", ")
	} else {
		results["feast"] = "Nothing to remove"
	}

	// 4. Wipe Outputs
	outputsDir := h.cfg.OutputsDir
	if _, err := os.Stat(outputsDir); err == nil {
		entries, _ := os.ReadDir(outputsDir)
		var removeErrors []string

		for _, entry := range entries {
			path := filepath.Join(outputsDir, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				removeErrors = append(removeErrors, entry.Name())
			}
		}

		if len(removeErrors) > 0 {
			results["outputs"] = "Partial: Failed to remove " + strings.Join(removeErrors, ", ")
		} else {
			results["outputs"] = "Wiped all files in outputs directory"
		}
	} else {
		os.MkdirAll(outputsDir, 0755)
		results["outputs"] = "Created missing outputs directory"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "System Reset Complete",
		"timestamp": time.Now().Format(time.RFC3339),
		"details":   results,
	})
}
