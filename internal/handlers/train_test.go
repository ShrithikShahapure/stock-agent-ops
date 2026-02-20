package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/handlers"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// ── TrainParent ───────────────────────────────────────────────────────────

func TestTrainParent_StartsTraining(t *testing.T) {
	cfg := config.Load()
	cfg.ParentDir = t.TempDir() // empty → no model file present

	mm := newMockManager()
	h := handlers.NewTrainHandler(cfg, mm)

	req := httptest.NewRequest(http.MethodPost, "/train-parent", nil)
	rec := httptest.NewRecorder()
	h.TrainParent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("TrainParent status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["task_id"] != "parent_training" {
		t.Errorf("TrainParent task_id = %v, want \"parent_training\"", resp["task_id"])
	}
}

func TestTrainParent_AlreadyRunning(t *testing.T) {
	cfg := config.Load()
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	mm.running["parent_training"] = true
	mm.statuses["parent_training"] = &tasks.TaskStatus{Status: "running"}

	h := handlers.NewTrainHandler(cfg, mm)

	req := httptest.NewRequest(http.MethodPost, "/train-parent", nil)
	rec := httptest.NewRecorder()
	h.TrainParent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("TrainParent(already running) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "already running" {
		t.Errorf("TrainParent(already running) status = %v, want \"already running\"", resp["status"])
	}
}

// ── TrainChild ────────────────────────────────────────────────────────────

func TestTrainChild_MissingTicker(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewTrainHandler(cfg, newMockManager())

	req := httptest.NewRequest(http.MethodPost, "/train-child",
		strings.NewReader(`{"ticker":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TrainChild(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("TrainChild(empty ticker) status = %d, want 400", rec.Code)
	}
}

func TestTrainChild_InvalidJSON(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewTrainHandler(cfg, newMockManager())

	req := httptest.NewRequest(http.MethodPost, "/train-child",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TrainChild(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("TrainChild(bad JSON) status = %d, want 400", rec.Code)
	}
}

func TestTrainChild_AlreadyRunning(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir()
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	mm.running["aapl"] = true
	mm.statuses["aapl"] = &tasks.TaskStatus{Status: "running"}
	// Simulate parent model completed so we skip that branch
	mm.statuses["parent_training"] = &tasks.TaskStatus{Status: "completed"}

	h := handlers.NewTrainHandler(cfg, mm)

	req := httptest.NewRequest(http.MethodPost, "/train-child",
		strings.NewReader(`{"ticker":"AAPL"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TrainChild(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("TrainChild(already running) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "running" {
		t.Errorf("TrainChild(already running) status = %v, want \"running\"", resp["status"])
	}
}

func TestTrainChild_StartsTraining(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir() // empty → no child model
	cfg.ParentDir = t.TempDir()  // empty → no parent model

	mm := newMockManager()
	// Simulate parent completed so TrainChild doesn't redirect to parent
	mm.statuses["parent_training"] = &tasks.TaskStatus{Status: "completed"}

	h := handlers.NewTrainHandler(cfg, mm)

	req := httptest.NewRequest(http.MethodPost, "/train-child",
		strings.NewReader(`{"ticker":"MSFT"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.TrainChild(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("TrainChild(start) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "started" {
		t.Errorf("TrainChild(start) status = %v, want \"started\"", resp["status"])
	}
	if resp["task_id"] != "msft" {
		t.Errorf("TrainChild(start) task_id = %v, want \"msft\"", resp["task_id"])
	}
}
