#!/bin/bash
# Smoke test script for stock-agent-ops API
# Usage: ./scripts/smoke.sh [base_url]

set -e

BASE_URL="${1:-http://localhost:8000}"
PASS=0
FAIL=0

echo "========================================"
echo "Stock Agent Ops - Smoke Test"
echo "Base URL: $BASE_URL"
echo "========================================"
echo ""

test_endpoint() {
    local method="$1"
    local path="$2"
    local data="$3"
    local expected="$4"
    local description="$5"

    echo -n "Testing $method $path - $description... "

    if [ "$method" = "GET" ]; then
        response=$(curl -sf "$BASE_URL$path" 2>/dev/null || echo "CURL_FAILED")
    elif [ "$method" = "POST" ]; then
        if [ -n "$data" ]; then
            response=$(curl -sf -X POST "$BASE_URL$path" -H "Content-Type: application/json" -d "$data" 2>/dev/null || echo "CURL_FAILED")
        else
            response=$(curl -sf -X POST "$BASE_URL$path" 2>/dev/null || echo "CURL_FAILED")
        fi
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -sf -X DELETE "$BASE_URL$path" 2>/dev/null || echo "CURL_FAILED")
    fi

    if [ "$response" = "CURL_FAILED" ]; then
        echo "FAILED (connection error)"
        FAIL=$((FAIL + 1))
        return 1
    fi

    if echo "$response" | grep -q "$expected"; then
        echo "PASSED"
        PASS=$((PASS + 1))
        return 0
    else
        echo "FAILED (expected '$expected')"
        echo "  Response: ${response:0:200}"
        FAIL=$((FAIL + 1))
        return 1
    fi
}

echo "--- Core Endpoints ---"
test_endpoint "GET" "/health" "" "healthy" "Health check"
test_endpoint "GET" "/" "" "project" "Root info"
test_endpoint "GET" "/openapi.json" "" "openapi" "OpenAPI spec"

echo ""
echo "--- Metrics ---"
echo -n "Testing GET /metrics - Prometheus metrics... "
metrics=$(curl -sf "$BASE_URL/metrics" 2>/dev/null || echo "CURL_FAILED")
if [ "$metrics" = "CURL_FAILED" ]; then
    echo "FAILED (connection error)"
    FAIL=$((FAIL + 1))
elif echo "$metrics" | grep -q "prediction_total\|system_cpu_percent"; then
    echo "PASSED"
    PASS=$((PASS + 1))
else
    echo "FAILED (missing expected metrics)"
    FAIL=$((FAIL + 1))
fi

echo ""
echo "--- System Endpoints ---"
test_endpoint "GET" "/system/cache" "" "cached_tickers\|count\|Redis" "Cache list"
test_endpoint "GET" "/outputs" "" "path\|contents\|total_items" "Outputs list"

echo ""
echo "--- Training Endpoints (non-destructive check) ---"
echo -n "Testing POST /train-parent - Parent training... "
response=$(curl -sf -X POST "$BASE_URL/train-parent" 2>/dev/null || echo "CURL_FAILED")
if [ "$response" = "CURL_FAILED" ]; then
    echo "FAILED (connection error)"
    FAIL=$((FAIL + 1))
elif echo "$response" | grep -qE "started|completed|running|rate"; then
    echo "PASSED"
    PASS=$((PASS + 1))
else
    echo "FAILED"
    FAIL=$((FAIL + 1))
fi

echo ""
echo "--- Prediction Endpoints ---"
echo -n "Testing POST /predict-child - Child prediction... "
response=$(curl -sf -X POST "$BASE_URL/predict-child" -H "Content-Type: application/json" -d '{"ticker":"AAPL"}' 2>/dev/null || echo "CURL_FAILED")
if [ "$response" = "CURL_FAILED" ]; then
    echo "FAILED (connection error)"
    FAIL=$((FAIL + 1))
elif echo "$response" | grep -qE "result|training|predictions"; then
    echo "PASSED"
    PASS=$((PASS + 1))
else
    echo "FAILED"
    echo "  Response: ${response:0:200}"
    FAIL=$((FAIL + 1))
fi

echo ""
echo "--- Status Endpoint ---"
echo -n "Testing GET /status/aapl - Task status... "
response=$(curl -s "$BASE_URL/status/aapl" 2>/dev/null || echo "CURL_FAILED")
if [ "$response" = "CURL_FAILED" ]; then
    echo "FAILED (connection error)"
    FAIL=$((FAIL + 1))
elif echo "$response" | grep -qE "status|not found|running|completed|failed"; then
    echo "PASSED"
    PASS=$((PASS + 1))
else
    echo "FAILED"
    echo "  Response: ${response:0:200}"
    FAIL=$((FAIL + 1))
fi

echo ""
echo "========================================"
echo "Results: $PASS passed, $FAIL failed"
echo "========================================"

if [ $FAIL -gt 0 ]; then
    exit 1
fi
exit 0
