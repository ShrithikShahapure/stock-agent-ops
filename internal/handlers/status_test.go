package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/handlers"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// chiRequest creates a request with chi URL params pre-populated.
func chiRequest(method, url string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestGetStatus_TaskNotFound(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir()
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	h := handlers.NewStatusHandler(cfg, mm)

	req := chiRequest(http.MethodGet, "/status/unknown", map[string]string{"task_id": "unknown"})
	rec := httptest.NewRecorder()
	h.GetStatus(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GetStatus(not found) status = %d, want 404", rec.Code)
	}
}

func TestGetStatus_RunningTask(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir()
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	mm.statuses["aapl"] = &tasks.TaskStatus{
		Status:    "running",
		StartTime: "2025-01-01 12:00:00",
	}

	h := handlers.NewStatusHandler(cfg, mm)

	req := chiRequest(http.MethodGet, "/status/aapl", map[string]string{"task_id": "aapl"})
	rec := httptest.NewRecorder()
	h.GetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetStatus(running) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "running" {
		t.Errorf("GetStatus(running) status = %v, want \"running\"", resp["status"])
	}
	if resp["elapsed_seconds"] == nil {
		t.Error("GetStatus(running) should include elapsed_seconds")
	}
}

func TestGetStatus_CompletedTask(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir()
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	mm.statuses["msft"] = &tasks.TaskStatus{
		Status:      "completed",
		CompletedAt: "2025-01-01 13:00:00",
	}

	h := handlers.NewStatusHandler(cfg, mm)

	req := chiRequest(http.MethodGet, "/status/msft", map[string]string{"task_id": "msft"})
	rec := httptest.NewRecorder()
	h.GetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetStatus(completed) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "completed" {
		t.Errorf("GetStatus(completed) status = %v, want \"completed\"", resp["status"])
	}
}

func TestGetStatus_ParentAliasMapping(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir()
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	// "parent" URL param should be mapped to "parent_training"
	mm.statuses["parent_training"] = &tasks.TaskStatus{Status: "completed"}

	h := handlers.NewStatusHandler(cfg, mm)

	req := chiRequest(http.MethodGet, "/status/parent", map[string]string{"task_id": "parent"})
	rec := httptest.NewRecorder()
	h.GetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetStatus(parent alias) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["task_id"] != "parent_training" {
		t.Errorf("GetStatus(parent alias) task_id = %v, want \"parent_training\"", resp["task_id"])
	}
}

func TestGetStatus_NilTaskManager_NoFile(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir()
	cfg.ParentDir = t.TempDir()

	// nil taskManager and no model file → 404
	h := handlers.NewStatusHandler(cfg, nil)

	req := chiRequest(http.MethodGet, "/status/ghost", map[string]string{"task_id": "ghost"})
	rec := httptest.NewRecorder()
	h.GetStatus(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GetStatus(nil manager, no file) status = %d, want 404", rec.Code)
	}
}
