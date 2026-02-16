# Checkpoint: Stock Agent Ops Fixes (2026-01-22)

## Original Issue
`Connection Failed: 'NoneType' object has no attribute 'get'` when running the stock analysis app.

---

## Root Causes Found & Fixed

### 1. JSON Output Corruption (Main Issue)
Python print statements and logging were going to stdout, corrupting the JSON output that Go expected.

**Files Fixed:**
| File | Change |
|------|--------|
| `logger/logger.py` | Changed `StreamHandler()` → `StreamHandler(sys.stderr)` |
| `src/data/ingestion.py` | All print statements → `file=sys.stderr` |
| `src/data/preparation.py` | Print statement → `file=sys.stderr` |
| `src/agents/graph.py` | Print statements → `file=sys.stderr` |

### 2. NoneType Safety Checks
Added null checks before calling `.get()` on API responses.

**Files Fixed:**
- `frontend/app.py:204-208` - Check if `data` is None before accessing
- `src/agents/graph.py` - Check if `raw_data` is None
- `src/agents/tools.py:60-62` - Check if `data` is None

### 3. Cached Null Values in Redis
Previous failed predictions cached null data. Cleared with:
```bash
docker-compose exec redis redis-cli KEYS "predict_child_*" | xargs redis-cli DEL
```

### 4. LLM Timeout Issues (5+ min analysis)
The multi-agent pipeline made 4 sequential LLM calls, too slow on CPU.

**Fix:** Rewrote `src/agents/graph.py` to use single LLM call:
- Before: 4 LLM calls (perf → news → report → critic) = 5+ minutes
- After: 1 LLM call = ~1 minute

### 5. Empty LLM Responses
Qwen3 model returned empty content with `SystemMessage`.

**Fix:** Changed to `HumanMessage` in `src/agents/graph.py`

### 6. LLM Connection Failures
llama.cpp container was crashing/unstable.

**Fixes in `src/agents/nodes.py`:**
- Added `invoke_llm_with_retry()` - 3 retries with 5s delay
- Lazy LLM client initialization with reconnection

**Fixes in `docker-compose.yml`:**
- Added `restart: unless-stopped` to llama service
- Increased memory limit to 10G
- Added `--parallel 2` for concurrent requests
- Added `--embeddings` flag

### 7. Frontend Timeout
Increased timeout in `frontend/app.py` from 120s → 300s

---

## Current Architecture

```
Frontend (Streamlit:8501)
    ↓ POST /analyze
API (Go:8000)
    ↓ calls Python CLI
scripts/ml_cli.py
    ↓
src/agents/graph.py (single LLM call)
    ├── fetch_prediction_data() → API /predict-child
    ├── get_stock_news() → Finnhub/Yahoo
    └── invoke_llm_with_retry() → llama.cpp:8080
```

---

## Services

| Service | URL | Port |
|---------|-----|------|
| Frontend | http://localhost:8501 | 8501 |
| API | http://localhost:8000 | 8000 |
| llama.cpp | http://localhost:8080 | 8080 |
| Redis | localhost | 6379 |
| Qdrant | localhost | 6333 |
| Prometheus | http://localhost:9090 | 9090 |
| Grafana | http://localhost:3000 | 3000 |

---

## Performance

- **Analysis time:** ~1 minute per stock (down from 5+ minutes)
- **First-time stock:** Triggers model training (~30-60s), then analysis
- **Cached results:** Returns instantly from Qdrant semantic cache

---

## Files Modified

```
frontend/app.py              # Null checks, timeout increase
logger/logger.py             # stderr for console handler
src/agents/graph.py          # Single LLM call, HumanMessage
src/agents/nodes.py          # Retry logic, lazy init
src/agents/tools.py          # Null checks
src/data/ingestion.py        # stderr for prints
src/data/preparation.py      # stderr for prints
docker-compose.yml           # llama stability improvements
```

---

## Commands

```bash
# Start all services
docker-compose up --build -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f api
docker-compose logs -f llama

# Clear prediction cache
docker-compose exec redis redis-cli KEYS "predict_child_*" | xargs redis-cli DEL

# Test endpoints
curl http://localhost:8000/health
curl -X POST http://localhost:8000/analyze -H "Content-Type: application/json" -d '{"ticker":"AAPL"}'
```
