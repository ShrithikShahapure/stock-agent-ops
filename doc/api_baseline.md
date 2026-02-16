# API Baseline Documentation

This document captures the expected request/response contracts for the FastAPI backend.
Use these as the source of truth when implementing the Go backend.

## Endpoints Summary

| Method | Path | Rate Limit | Description |
|--------|------|------------|-------------|
| GET | `/` | - | Project info |
| GET | `/health` | - | Health check |
| GET | `/docs` | - | Swagger UI |
| GET | `/openapi.json` | - | OpenAPI spec |
| GET | `/metrics` | - | Prometheus metrics |
| POST | `/analyze` | - | AI agent analysis |
| POST | `/train-parent` | 5/hour | Train parent model |
| POST | `/train-child` | 5/hour | Train child model |
| POST | `/predict-parent` | 40/hour | Parent prediction |
| POST | `/predict-child` | 40/hour | Child prediction |
| GET | `/status/{task_id}` | - | Task status |
| POST | `/monitor/parent` | - | Parent drift + eval |
| POST | `/monitor/{ticker}` | - | Ticker drift + eval |
| GET | `/monitor/{ticker}/drift` | - | Drift report |
| GET | `/monitor/{ticker}/eval` | - | Eval report |
| GET | `/system/logs` | - | Latest logs |
| GET | `/system/cache` | - | Redis cache |
| DELETE | `/system/reset` | - | System reset |
| GET | `/outputs` | - | Output files list |
| GET | `/outputs/{ticker}` | - | Ticker outputs |

---

## Endpoint Contracts

### GET /health

**Response (200)**:
```json
{
  "status": "healthy"
}
```

---

### GET /

**Response (200)**:
```json
{
  "project": "MLOps Stock Prediction Pipeline",
  "version": "3.1",
  "description": "Production-ready MLOps system for stock price prediction using LSTM and Transfer Learning",
  "features": [
    "Parent-Child Transfer Learning Strategy",
    "Real-time predictions with Redis caching",
    "Feast feature store integration",
    "MLflow experiment tracking",
    "Qdrant semantic memory for AI agents",
    "Prometheus monitoring & Grafana dashboards",
    "Auto-healing: Missing models trigger background training"
  ],
  "endpoints": {
    "health": "GET /health - Health check",
    "docs": "GET /docs - Interactive API documentation",
    "training": {
      "train_parent": "POST /train-parent - Train parent model (S&P 500)",
      "train_child": "POST /train-child - Train child model for specific ticker"
    },
    "prediction": {
      "predict_parent": "POST /predict-parent - Predict using parent model",
      "predict_child": "POST /predict-child - Predict using child model (auto-trains if missing)"
    },
    "monitoring": {
      "status": "GET /status/{task_id} - Check training task status",
      "monitor_parent": "POST /monitor/parent - Monitor parent model drift & agent eval",
      "monitor_ticker": "POST /monitor/{ticker} - Monitor specific ticker",
      "drift_report": "GET /monitor/{ticker}/drift - Get drift analysis JSON",
      "eval_report": "GET /monitor/{ticker}/eval - Get agent evaluation JSON"
    },
    "system": {
      "outputs": "GET /outputs - List all files in outputs directory",
      "cache": "GET /system/cache - Inspect Redis cache",
      "logs": "GET /system/logs - Retrieve latest log lines",
      "reset": "DELETE /system/reset - Wipe all system data (Redis, Qdrant, Feast, Outputs)",
      "metrics": "GET /metrics - Prometheus metrics"
    },
    "agent": {
      "analyze": "POST /analyze - Analyze stock with AI agent"
    }
  },
  "quick_start": {
    "1_train_parent": "curl -X POST http://localhost:8000/train-parent",
    "2_predict_child": "curl -X POST http://localhost:8000/predict-child -H 'Content-Type: application/json' -d '{\"ticker\": \"AAPL\"}'",
    "3_check_status": "curl -X GET http://localhost:8000/status/aapl",
    "4_view_outputs": "curl -X GET http://localhost:8000/outputs"
  },
  "documentation": "See /docs for full interactive API documentation"
}
```

---

### POST /analyze

**Request**:
```json
{
  "ticker": "AAPL",
  "use_fmi": false,
  "thread_id": "optional-thread-id"
}
```

**Response (200)** - Success:
```json
{
  "final_report": "...",
  "recommendation": "BUY/SELL/HOLD",
  "confidence": 0.85,
  "last_price": 150.25,
  "predictions": {...}
}
```

**Response (500)** - Error:
```json
{
  "detail": "Analysis failed: <error message>"
}
```

---

### POST /train-parent

**Request**: None (empty body)

**Response (200)** - Started:
```json
{
  "status": "started",
  "task_id": "parent_training"
}
```

**Response (200)** - Already exists:
```json
{
  "status": "completed",
  "task_id": "parent_training",
  "detail": "Parent model already exists"
}
```

**Response (200)** - Already running:
```json
{
  "status": "already running",
  "task_id": "parent_training"
}
```

**Response (429)** - Rate limited:
```json
{
  "detail": "Rate limit exceeded"
}
```

---

### POST /train-child

**Request**:
```json
{
  "ticker": "AAPL"
}
```

**Response (200)** - Started:
```json
{
  "status": "started",
  "task_id": "aapl"
}
```

**Response (200)** - Completed (model exists):
```json
{
  "status": "completed",
  "task_id": "aapl",
  "detail": "Model already exists"
}
```

**Response (200)** - Running:
```json
{
  "status": "running",
  "task_id": "aapl",
  "detail": "Training already in progress"
}
```

**Response (200)** - Parent training triggered:
```json
{
  "status": "started_parent",
  "task_id": "parent_training",
  "detail": "Parent model missing. Training parent first."
}
```

**Response (400)** - Missing ticker:
```json
{
  "detail": "ticker is required"
}
```

---

### POST /predict-parent

**Request**: None (empty body)

**Response (200)**:
```json
{
  "result": {
    "ticker": "^GSPC",
    "predictions": [...],
    "last_price": 4500.25,
    "timestamp": "2025-01-19T12:00:00"
  }
}
```

---

### POST /predict-child

**Request**:
```json
{
  "ticker": "AAPL"
}
```

**Response (200)** - Success:
```json
{
  "result": {
    "ticker": "AAPL",
    "predictions": [...],
    "last_price": 150.25,
    "timestamp": "2025-01-19T12:00:00"
  }
}
```

**Response (202)** - Training started (model missing):
```json
{
  "status": "training",
  "detail": "Model for AAPL missing. Training started (with auto-prediction).",
  "task_id": "aapl"
}
```

**Response (202)** - Training in progress:
```json
{
  "status": "training",
  "detail": "Training in progress. Please retry later.",
  "task_id": "aapl"
}
```

**Response (202)** - Parent training triggered:
```json
{
  "status": "training",
  "detail": "Parent model missing. Training parent first.",
  "task_id": "parent_training"
}
```

---

### GET /status/{task_id}

**Path Parameters**:
- `task_id`: Task ID (e.g., "aapl", "parent", "parent_training")
- Note: "parent" is mapped to "parent_training"

**Response (200)** - Running:
```json
{
  "status": "running",
  "elapsed_seconds": 45,
  "task_id": "aapl"
}
```

**Response (200)** - Completed:
```json
{
  "status": "completed",
  "result": {...},
  "completed_at": "2025-01-19 12:00:00",
  "task_id": "aapl"
}
```

**Response (200)** - Failed:
```json
{
  "status": "failed",
  "error": "Error message",
  "failed_at": "2025-01-19 12:00:00",
  "task_id": "aapl"
}
```

**Response (200)** - Model found on disk:
```json
{
  "status": "completed",
  "detail": "Model file found on disk",
  "task_id": "aapl"
}
```

**Response (404)** - Not found:
```json
{
  "detail": "Task 'aapl' not found."
}
```

---

### GET /system/logs

**Query Parameters**:
- `lines`: Number of lines (default: 100)

**Response (200)**:
```json
{
  "logs": "2025-01-19 12:00:00 - INFO - ...\n...",
  "filename": "pipeline_20250119_120000.log"
}
```

**Response (200)** - No logs:
```json
{
  "logs": "No log files found."
}
```

---

### POST /monitor/parent

**Response (200)**:
```json
{
  "ticker": "^GSPC",
  "type": "Parent Model (Market Index)",
  "drift": {
    "status": "passed|failed",
    "health": "healthy|degraded|critical",
    ...
  },
  "agent_eval": {
    "status": "passed|failed",
    "scores": {...}
  },
  "links": {
    "get_drift_json": "/monitor/^GSPC/drift",
    "get_eval_json": "/monitor/^GSPC/eval"
  }
}
```

---

### POST /monitor/{ticker}

**Response (200)**:
```json
{
  "ticker": "AAPL",
  "is_parent": false,
  "drift": {
    "status": "skipped",
    "detail": "Drift calculation reserved for parent model."
  },
  "agent_eval": {
    "status": "passed|failed",
    ...
  }
}
```

---

### GET /monitor/{ticker}/drift

**Response (200)** - JSON report:
```json
{
  "ticker": "^gspc",
  "timestamp": "2025-01-19T12:00:00",
  "health": "healthy",
  "metrics": {...}
}
```

**Response (200)** - Files only (no JSON):
```json
{
  "files": ["drift_report.html", "..."],
  "message": "Access HTML report in outputs/",
  "detail": "JSON summary missing."
}
```

**Response (404)**:
```json
{
  "detail": "No drift report found"
}
```

---

### GET /monitor/{ticker}/eval

**Response (200)**:
```json
{
  "ticker": "AAPL",
  "timestamp": "2025-01-19T12:00:00",
  "scores": {...}
}
```

**Response (404)**:
```json
{
  "detail": "No evaluation found. Run POST /monitor/{ticker} first."
}
```

---

### DELETE /system/reset

**Response (200)**:
```json
{
  "status": "System Reset Complete",
  "timestamp": "2025-01-19T12:00:00.000000",
  "details": {
    "redis": "✅ Flushed",
    "qdrant": "✅ Collection wiped and recreated",
    "feast": "✅ Removed: registry.db, features.parquet",
    "outputs": "✅ Wiped all files in outputs directory"
  }
}
```

---

### GET /system/cache

**Query Parameters**:
- `ticker`: Optional ticker to get specific cache

**Response (200)** - List cached:
```json
{
  "cached_tickers": ["AAPL", "GOOG", "TSLA"],
  "count": 3
}
```

**Response (200)** - Specific ticker:
```json
{
  "ticker": "AAPL",
  "predictions": [...],
  ...
}
```

**Response (404)**:
```json
{
  "detail": "No cache found for AAPL"
}
```

**Response (503)**:
```json
{
  "detail": "Redis not connected"
}
```

---

### GET /outputs

**Response (200)**:
```json
{
  "path": "outputs",
  "total_items": 5,
  "directories": 3,
  "files": 2,
  "total_size_kb": 1024.5,
  "contents": [
    {
      "name": "parent",
      "type": "directory",
      "path": "parent",
      "file_count": 10
    },
    {
      "name": "aapl",
      "type": "directory",
      "path": "aapl",
      "file_count": 8
    }
  ],
  "note": "Use GET /outputs/{ticker} to see detailed contents of a specific ticker directory"
}
```

---

### GET /outputs/{ticker}

**Response (200)**:
```json
{
  "ticker": "AAPL",
  "path": "outputs/aapl",
  "total_files": 8,
  "total_size_kb": 512.25,
  "categories": ["drift", "agent_eval", "models"],
  "files_by_category": {
    "drift": [
      {
        "name": "latest_drift.json",
        "path": "drift/latest_drift.json",
        "size_kb": 2.5,
        "modified": "2025-01-19 12:00:00",
        "category": "drift"
      }
    ],
    ...
  },
  "all_files": [...]
}
```

**Response (404)**:
```json
{
  "detail": "No outputs found for ticker 'XYZ'"
}
```

---

## Redis Key Formats

| Key Pattern | Purpose | TTL |
|-------------|---------|-----|
| `task_status:{task_id}` | Training task status | 2h (running), 1h (completed/failed) |
| `predict_child_{ticker}` | Prediction cache | 24h (86400s) |
| `rate_limit:{prefix}:{window}` | Rate limit counter | window_sec |

---

## Prometheus Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `system_cpu_percent` | Gauge | - | CPU usage % |
| `system_ram_used_mb` | Gauge | - | RAM used in MB |
| `system_disk_used_mb` | Gauge | - | Disk used in MB |
| `redis_up` | Gauge | - | Redis status (1=up, 0=down) |
| `redis_keys_total` | Gauge | - | Total Redis keys |
| `training_status` | Gauge | `task_id` | 0=idle, 1=running, 2=completed |
| `training_mse_last` | Gauge | - | Last training MSE |
| `training_duration_seconds` | Histogram | `task_id` | Training duration |
| `prediction_total` | Counter | `type` | Total predictions (parent/child) |
| `prediction_latency_seconds` | Histogram | `type` | Prediction latency |
| `redis_cache_hit_total` | Counter | `key` | Cache hits |
| `redis_cache_miss_total` | Counter | `key` | Cache misses |

---

## Auto-Healing Behavior

### /predict-child Flow

1. Check Redis cache for `predict_child_{ticker}`
   - If HIT: return cached result
   - If MISS: continue

2. Execute `predict_child(ticker)`
   - If SUCCESS: cache result, return
   - If MODEL_MISSING: continue to auto-heal

3. Check if parent model exists
   - If MISSING: trigger `train_parent`, return 202
   - If EXISTS: continue

4. Trigger `train_child(ticker)` with chain function
   - Chain function: after training, cache prediction
   - Return 202 with status "training"

### /train-child Flow

1. Check if parent model exists
   - If MISSING and not training: trigger `train_parent`
   - Return "started_parent" status

2. Check if child model exists
   - If EXISTS: return "completed"

3. Check if training already running
   - If RUNNING: return "running"

4. Start training with chain function
   - Chain function: auto-predict and cache after training
   - Return "started"
