package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shrithkshahapure/stock-agent-ops/internal/handlers"
	"github.com/shrithkshahapure/stock-agent-ops/internal/services/python"
)

// mockRunner is a test double that implements python.RunnerInterface.
type mockRunner struct {
	analyzeResult *python.Result
	analyzeErr    error
	predictResult *python.Result
	predictErr    error
}

func (m *mockRunner) TrainParent(_ context.Context) (*python.Result, error) {
	return &python.Result{Data: map[string]interface{}{"status": "ok"}}, nil
}
func (m *mockRunner) TrainChild(_ context.Context, _ string) (*python.Result, error) {
	return &python.Result{Data: map[string]interface{}{"status": "ok"}}, nil
}
func (m *mockRunner) PredictParent(_ context.Context) (*python.Result, error) {
	if m.predictErr != nil {
		return nil, m.predictErr
	}
	return m.predictResult, nil
}
func (m *mockRunner) PredictChild(_ context.Context, _ string) (*python.Result, error) {
	if m.predictErr != nil {
		return nil, m.predictErr
	}
	return m.predictResult, nil
}
func (m *mockRunner) Analyze(_ context.Context, _, _ string) (*python.Result, error) {
	return m.analyzeResult, m.analyzeErr
}
func (m *mockRunner) MonitorParent(_ context.Context) (*python.Result, error) {
	return &python.Result{Data: map[string]interface{}{}}, nil
}
func (m *mockRunner) MonitorTicker(_ context.Context, _ string) (*python.Result, error) {
	return &python.Result{Data: map[string]interface{}{}}, nil
}

func TestAnalyze_MissingTicker(t *testing.T) {
	h := handlers.NewAnalyzeHandler(&mockRunner{})

	body := `{"ticker": ""}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Analyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Analyze(empty ticker) status = %d, want 400", rec.Code)
	}
}

func TestAnalyze_InvalidJSON(t *testing.T) {
	h := handlers.NewAnalyzeHandler(&mockRunner{})

	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Analyze(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Analyze(bad JSON) status = %d, want 400", rec.Code)
	}
}

func TestAnalyze_RunnerError(t *testing.T) {
	h := handlers.NewAnalyzeHandler(&mockRunner{
		analyzeErr: errors.New("python crashed"),
	})

	body := `{"ticker": "AAPL"}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Analyze(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Analyze(runner error) status = %d, want 500", rec.Code)
	}
	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["detail"] == "" {
		t.Error("Analyze(runner error) response should include \"detail\"")
	}
}

func TestAnalyze_Success(t *testing.T) {
	h := handlers.NewAnalyzeHandler(&mockRunner{
		analyzeResult: &python.Result{
			Data: map[string]interface{}{
				"final_report":   "AAPL looks bullish",
				"recommendation": "BULLISH",
				"confidence":     "High",
			},
		},
	})

	body := `{"ticker": "AAPL", "thread_id": "test-thread"}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Analyze(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Analyze(success) status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Analyze(success) body is not valid JSON: %v", err)
	}
	if resp["recommendation"] != "BULLISH" {
		t.Errorf("Analyze(success) recommendation = %v, want BULLISH", resp["recommendation"])
	}
}
