package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
)

// HealthHandler handles health-related endpoints
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// Root handles GET /
func (h *HealthHandler) Root(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"project":     "MLOps Stock Prediction Pipeline",
		"version":     "3.1",
		"description": "Production-ready MLOps system for stock price prediction using LSTM and Transfer Learning",
		"features": []string{
			"Parent-Child Transfer Learning Strategy",
			"Real-time predictions with Redis caching",
			"Feast feature store integration",
			"MLflow experiment tracking",
			"Qdrant semantic memory for AI agents",
			"Prometheus monitoring & Grafana dashboards",
			"Auto-healing: Missing models trigger background training",
		},
		"endpoints": map[string]interface{}{
			"health": "GET /health - Health check",
			"docs":   "GET /docs - Interactive API documentation",
			"training": map[string]string{
				"train_parent": "POST /train-parent - Train parent model (S&P 500)",
				"train_child":  "POST /train-child - Train child model for specific ticker",
			},
			"prediction": map[string]string{
				"predict_parent": "POST /predict-parent - Predict using parent model",
				"predict_child":  "POST /predict-child - Predict using child model (auto-trains if missing)",
			},
			"monitoring": map[string]string{
				"status":         "GET /status/{task_id} - Check training task status",
				"monitor_parent": "POST /monitor/parent - Monitor parent model drift & agent eval",
				"monitor_ticker": "POST /monitor/{ticker} - Monitor specific ticker",
				"drift_report":   "GET /monitor/{ticker}/drift - Get drift analysis JSON",
				"eval_report":    "GET /monitor/{ticker}/eval - Get agent evaluation JSON",
			},
			"system": map[string]string{
				"outputs": "GET /outputs - List all files in outputs directory",
				"cache":   "GET /system/cache - Inspect Redis cache",
				"logs":    "GET /system/logs - Retrieve latest log lines",
				"reset":   "DELETE /system/reset - Wipe all system data (Redis, Qdrant, Feast, Outputs)",
				"metrics": "GET /metrics - Prometheus metrics",
			},
			"agent": map[string]string{
				"analyze": "POST /analyze - Analyze stock with AI agent",
			},
		},
		"quick_start": map[string]string{
			"1_train_parent":  "curl -X POST http://localhost:8000/train-parent",
			"2_predict_child": "curl -X POST http://localhost:8000/predict-child -H 'Content-Type: application/json' -d '{\"ticker\": \"AAPL\"}'",
			"3_check_status":  "curl -X GET http://localhost:8000/status/aapl",
			"4_view_outputs":  "curl -X GET http://localhost:8000/outputs",
		},
		"documentation": "See /docs for full interactive API documentation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// OpenAPI serves the OpenAPI spec
func (h *HealthHandler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	// Try to serve from doc/openapi.json if it exists
	specPath := filepath.Join("doc", "openapi.json")
	if _, err := os.Stat(specPath); err == nil {
		http.ServeFile(w, r, specPath)
		return
	}

	// Return a minimal OpenAPI spec
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Stock Agent Ops API",
			"version":     "3.1",
			"description": "MLOps Stock Prediction Pipeline API",
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health check",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
						},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

// Docs serves the Swagger UI
func (h *HealthHandler) Docs(w http.ResponseWriter, r *http.Request) {
	// Serve a simple HTML page that loads Swagger UI from CDN
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Stock Agent Ops API</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout"
            });
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
