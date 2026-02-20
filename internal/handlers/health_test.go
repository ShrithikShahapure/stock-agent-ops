package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/handlers"
)

func TestHealth_Returns200WithHealthyStatus(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Health() status = %d, want 200", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Health() body is not valid JSON: %v", err)
	}
	if resp["status"] != "healthy" {
		t.Errorf("Health() status = %q, want \"healthy\"", resp["status"])
	}
}

func TestRoot_ContainsRequiredKeys(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.Root(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Root() status = %d, want 200", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Root() body is not valid JSON: %v", err)
	}

	for _, key := range []string{"project", "version", "endpoints", "quick_start"} {
		if _, ok := resp[key]; !ok {
			t.Errorf("Root() response missing key %q", key)
		}
	}
}

func TestOpenAPI_ReturnsFallbackSpecWhenFileAbsent(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	h.OpenAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("OpenAPI() status = %d, want 200", rec.Code)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("OpenAPI() body is not valid JSON: %v", err)
	}
	if spec["openapi"] == nil {
		t.Error("OpenAPI() response missing \"openapi\" field")
	}
}

func TestDocs_ReturnsHTML(t *testing.T) {
	cfg := config.Load()
	h := handlers.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	h.Docs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Docs() status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html" {
		t.Errorf("Docs() Content-Type = %q, want \"text/html\"", ct)
	}
}
