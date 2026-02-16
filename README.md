# Stock Agent Ops

End-to-end stock market prediction and analysis system using transfer learning (LSTM) and agentic AI (LangGraph).

A Go API orchestrates Python ML pipelines, serves predictions with Redis caching, and generates Bloomberg-style analysis reports via a local LLM (Qwen3 7B on llama.cpp).

---

## Architecture

```
User Layer              Logic Layer                       Storage Layer
────────────           ────────────────────────           ──────────────
Streamlit UI  ──────>  Go API (Chi router)  ──────────>  Redis (cache, tasks, rate limits)
  :8501                  |                                Qdrant (semantic vector cache)
Monitoring UI ──────>    |── Python ML CLI (subprocess)   Feast (feature store)
  :8502                  |── llama.cpp (Qwen3 7B)         MLflow/DagsHub (experiment tracking)
                         |── Prometheus ──> Grafana        Filesystem (models, logs, outputs)
```

**How it works:**
1. The **Go backend** handles all HTTP requests, rate limiting, caching, and async task management
2. ML operations (training, prediction, analysis) are delegated to **Python** via a CLI wrapper (`scripts/ml_cli.py`)
3. The **LSTM model** is trained on S&P 500 (parent), then fine-tuned per ticker via transfer learning (child)
4. **LangGraph agents** call the LLM to generate analysis reports, cached in Qdrant for 24h

---

## Tech Stack

| Component | Technology |
|:---|:---|
| Backend | Go (Chi router) |
| ML Models | PyTorch LSTM (transfer learning) |
| LLM | llama.cpp (Qwen3 7B, quantized GGUF) |
| AI Agents | LangGraph + LangChain |
| Feature Store | Feast (offline: Parquet, online: Redis) |
| Experiment Tracking | MLflow (via DagsHub) |
| Semantic Cache | Qdrant (768-dim vectors, 24h TTL) |
| Prediction Cache | Redis (24h TTL) |
| Monitoring | Prometheus + Grafana |
| Frontend | Streamlit |
| Deployment | Docker Compose / Kubernetes (Minikube) |

---

## Quick Start

### Prerequisites

- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Finnhub API key](https://finnhub.io/) (free tier works)

### 1. Clone

```bash
git clone https://github.com/ShrithikShahapure/stock-agent-ops.git
cd stock-agent-ops
```

### 2. Download LLM Model

```bash
mkdir -p models

# Qwen3 7B (recommended, ~4.4 GB)
wget -O models/qwen3-7b-q4_k_m.gguf \
  https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/resolve/main/qwen2.5-7b-instruct-q4_k_m.gguf
```

### 3. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and add at minimum:

```
FMI_API_KEY=your_finnhub_api_key
```

Optional (for remote experiment tracking):
```
DAGSHUB_USER_NAME=
DAGSHUB_REPO_NAME=
DAGSHUB_TOKEN=
MLFLOW_TRACKING_URI=
```

### 4. Start

```bash
./run_docker.sh
```

Or directly:

```bash
docker-compose up --build -d
```

### 5. Access

| Service | URL | Credentials |
|:---|:---|:---|
| Streamlit UI | http://localhost:8501 | - |
| Monitoring Dashboard | http://localhost:8502 | - |
| API Docs (Swagger) | http://localhost:8000/docs | - |
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| llama.cpp | http://localhost:8080/v1 | - |

---

## API Endpoints

### Training

```bash
# Train parent model (S&P 500)
curl -X POST http://localhost:8000/train-parent

# Train child model (transfer learning from parent)
curl -X POST http://localhost:8000/train-child \
  -H "Content-Type: application/json" \
  -d '{"ticker":"AAPL"}'

# Check training status
curl http://localhost:8000/status/aapl
```

### Prediction

```bash
# Predict with child model (auto-trains if model missing)
curl -X POST http://localhost:8000/predict-child \
  -H "Content-Type: application/json" \
  -d '{"ticker":"AAPL"}'

# Predict with parent model
curl -X POST http://localhost:8000/predict-parent
```

### Analysis (LLM Agent)

```bash
# Full Bloomberg-style analysis
curl -X POST http://localhost:8000/analyze \
  -H "Content-Type: application/json" \
  -d '{"ticker":"AAPL"}'
```

### Monitoring

```bash
# Drift detection + agent evaluation (parent)
curl -X POST http://localhost:8000/monitor/parent

# Per-ticker monitoring
curl -X POST http://localhost:8000/monitor/AAPL

# View drift report
curl http://localhost:8000/monitor/AAPL/drift

# View agent evaluation
curl http://localhost:8000/monitor/AAPL/eval
```

### System

```bash
curl http://localhost:8000/health
curl http://localhost:8000/outputs
curl http://localhost:8000/system/logs?lines=50
curl http://localhost:8000/system/cache
curl http://localhost:8000/metrics
curl -X DELETE http://localhost:8000/system/reset
```

Rate limits: 5/hour for training, 40/hour for predictions.

---

## How Transfer Learning Works

```
Parent Model (S&P 500)                Child Model (e.g. AAPL)
─────────────────────                 ──────────────────────
^GSPC data (2004-present)    ──>      Load parent weights
Train LSTM from scratch               Freeze LSTM layers (or fine-tune all)
20 epochs                              Train FC layer on ticker data
Save to outputs/parent/                10 epochs, save to outputs/AAPL/
```

The parent model learns general market patterns from the S&P 500 index. Child models inherit these patterns and specialize on individual tickers, requiring less data and fewer epochs.

**Model architecture:** 3-layer LSTM (128 hidden, 20% dropout) with 7 input features (Open, High, Low, Close, Volume, RSI14, MACD). Context window: 60 days. Forecast horizon: 5 business days.

---

## How AI Analysis Works

When you call `/analyze`, the system:

1. **Checks Qdrant** for a cached analysis (similarity > 0.95, same ticker, < 24h old)
2. **Fetches predictions** from `/predict-child` (trains model if missing)
3. **Fetches news** from Finnhub (falls back to Yahoo Finance)
4. **Calls LLM once** with predictions + news to generate a Bloomberg-style report
5. **Caches the result** in Qdrant with embeddings for future lookups

The report includes a market stance (BULLISH/BEARISH/NEUTRAL) and confidence level.

---

## Project Structure

```
cmd/server/                  Go entrypoint
internal/
  config/                    Environment-based configuration
  handlers/                  HTTP handlers (health, train, predict, analyze, monitor, system, outputs)
  http/                      Chi router setup
  metrics/                   Prometheus metrics (matches Grafana dashboards)
  middleware/                CORS, logging, rate limiting, panic recovery
  models/                    Request/response structs
  services/
    cache/                   Redis prediction cache (24h TTL)
    python/                  Python CLI subprocess runner
    redis/                   Redis client wrapper
    tasks/                   Background task manager (max 4 workers)

src/
  agents/                    LangGraph agent (graph.py), nodes, tools
  data/                      Data ingestion (yfinance) and preparation (PyTorch datasets)
  llm/                       LLM provider abstraction (llama.cpp primary, Ollama fallback)
  memory/                    Qdrant semantic cache
  model/                     LSTM definition, training, evaluation, saving
  monitoring/                Drift detection, agent evaluation
  pipelines/                 Training pipeline (parent/child), inference pipeline

scripts/
  ml_cli.py                  Python CLI called by Go (train, predict, analyze, monitor)
  smoke.sh                   Endpoint smoke tests

frontend/                    Streamlit UI (port 8501)
monitoring_app/              Monitoring Streamlit UI (port 8502)
feature_store/               Feast config and feature definitions
k8s/                         Kubernetes manifests (api, redis, qdrant, llama, prometheus, grafana, frontends)
prometheus/                  Prometheus scrape config
doc/                         System design, API baseline, deployment docs
```

---

## Kubernetes Deployment

```bash
./run_k8s.sh
```

This starts Minikube, builds images, deploys all services, and waits for readiness. Run `sudo minikube tunnel` in a separate terminal for LoadBalancer access.

---

## Smoke Tests

```bash
./scripts/smoke.sh              # defaults to localhost:8000
./scripts/smoke.sh http://your-host:8000
```

---

## Environment Variables

| Variable | Default | Description |
|:---|:---|:---|
| `PORT` | `8000` | API server port |
| `REDIS_HOST` | `localhost:6379` | Redis address |
| `QDRANT_HOST` | `qdrant` | Qdrant hostname |
| `LLAMA_CPP_BASE_URL` | - | llama.cpp server URL (e.g. `http://llama:8080/v1`) |
| `LLM_MODEL` | `qwen3-7b` | Model name for llama.cpp |
| `FMI_API_KEY` | - | Finnhub API key for news |
| `MLFLOW_TRACKING_URI` | - | MLflow tracking server (optional) |
| `DAGSHUB_USER_NAME` | - | DagsHub username (optional) |
| `DAGSHUB_REPO_NAME` | - | DagsHub repo name (optional) |
| `DAGSHUB_TOKEN` | - | DagsHub token (optional) |

---

## License

MIT
