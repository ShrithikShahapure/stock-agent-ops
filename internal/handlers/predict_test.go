package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/handlers"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/tasks"
)

// mockCache is a test double that implements cache.CacheInterface.
type mockCache struct {
	data    map[string]map[string]interface{}
	tickers []string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]map[string]interface{})}
}

func (c *mockCache) Get(ticker string) (map[string]interface{}, bool) {
	v, ok := c.data[strings.ToLower(ticker)]
	return v, ok
}
func (c *mockCache) Set(ticker string, data map[string]interface{}) error {
	c.data[strings.ToLower(ticker)] = data
	return nil
}
func (c *mockCache) Delete(ticker string) error {
	delete(c.data, strings.ToLower(ticker))
	return nil
}
func (c *mockCache) GetCachedTickers() ([]string, error) { return c.tickers, nil }
func (c *mockCache) GetForTicker(ticker string) (map[string]interface{}, error) {
	v, ok := c.data[strings.ToLower(ticker)]
	if !ok {
		return nil, nil
	}
	return v, nil
}

// mockManager is a test double that implements tasks.ManagerInterface.
type mockManager struct {
	statuses map[string]*tasks.TaskStatus
	running  map[string]bool
}

func newMockManager() *mockManager {
	return &mockManager{
		statuses: make(map[string]*tasks.TaskStatus),
		running:  make(map[string]bool),
	}
}

func (m *mockManager) GetStatus(taskID string) *tasks.TaskStatus { return m.statuses[taskID] }
func (m *mockManager) IsRunning(taskID string) bool              { return m.running[taskID] }
func (m *mockManager) StartTrainParent() (bool, error)           { return true, nil }
func (m *mockManager) StartTrainChild(_ string, _ func()) (bool, error) {
	return true, nil
}

// ── PredictParent ──────────────────────────────────────────────────────────

func TestPredictParent_RunnerError(t *testing.T) {
	cfg := config.Load()
	cfg.ParentDir = t.TempDir()

	h := handlers.NewPredictHandler(cfg,
		&mockRunner{predictErr: errors.New("model not found")},
		nil, nil, nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/predict-parent", nil)
	rec := httptest.NewRecorder()
	h.PredictParent(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("PredictParent(runner error) status = %d, want 500", rec.Code)
	}
}

func TestPredictParent_Success(t *testing.T) {
	cfg := config.Load()
	cfg.ParentDir = t.TempDir()

	h := handlers.NewPredictHandler(cfg,
		&mockRunner{predictResult: &python.Result{Data: map[string]interface{}{
			"ticker": "^GSPC",
		}}},
		nil, nil, nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/predict-parent", nil)
	rec := httptest.NewRecorder()
	h.PredictParent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PredictParent(success) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["result"] == nil {
		t.Error("PredictParent(success) response missing \"result\" key")
	}
}

// ── PredictChild ───────────────────────────────────────────────────────────

func TestPredictChild_MissingTicker(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewPredictHandler(cfg, &mockRunner{}, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/predict-child",
		strings.NewReader(`{"ticker":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.PredictChild(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PredictChild(empty ticker) status = %d, want 400", rec.Code)
	}
}

func TestPredictChild_InvalidJSON(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewPredictHandler(cfg, &mockRunner{}, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/predict-child",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.PredictChild(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PredictChild(bad JSON) status = %d, want 400", rec.Code)
	}
}

func TestPredictChild_CacheHit(t *testing.T) {
	cfg := config.Load()
	mc := newMockCache()
	mc.data["aapl"] = map[string]interface{}{"ticker": "AAPL", "cached": true}

	h := handlers.NewPredictHandler(cfg, &mockRunner{}, mc, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/predict-child",
		strings.NewReader(`{"ticker":"AAPL"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.PredictChild(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PredictChild(cache hit) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	result, _ := resp["result"].(map[string]interface{})
	if result["cached"] != true {
		t.Error("PredictChild(cache hit) expected cached data in response")
	}
}

func TestPredictChild_Success_CachesResult(t *testing.T) {
	cfg := config.Load()
	mc := newMockCache()
	h := handlers.NewPredictHandler(cfg,
		&mockRunner{predictResult: &python.Result{Data: map[string]interface{}{
			"ticker": "TSLA",
		}}},
		mc, nil, nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/predict-child",
		strings.NewReader(`{"ticker":"TSLA"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.PredictChild(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PredictChild(success) status = %d, want 200", rec.Code)
	}
	// Verify the result was stored in the cache
	if _, found := mc.Get("TSLA"); !found {
		t.Error("PredictChild(success) result should be cached after prediction")
	}
}

func TestPredictChild_ModelMissing_StartsTraining(t *testing.T) {
	cfg := config.Load()
	cfg.OutputsDir = t.TempDir() // empty dir → no model files
	cfg.ParentDir = t.TempDir()

	mm := newMockManager()
	h := handlers.NewPredictHandler(cfg,
		&mockRunner{predictErr: errors.New("missing model file")},
		newMockCache(), mm, nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/predict-child",
		strings.NewReader(`{"ticker":"NVDA"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.PredictChild(rec, req)

	// Should return 202 Accepted (training started)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("PredictChild(missing model) status = %d, want 202", rec.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "training" {
		t.Errorf("PredictChild(missing model) status = %v, want \"training\"", resp["status"])
	}
}
