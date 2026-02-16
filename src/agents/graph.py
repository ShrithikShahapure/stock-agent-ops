"""
Fast stock analysis with single LLM call.
Replaces slow multi-agent pipeline for better performance on CPU.
"""
import sys
from langchain_core.messages import HumanMessage

from src.agents.tools import get_stock_news, fetch_prediction_data
from src.agents.nodes import invoke_llm_with_retry
from src.memory.semantic_cache import SemanticCache
from src.llm.embeddings import get_embeddings_client
from logger.logger import get_logger

logger = get_logger()


def analyze_stock(ticker: str, thread_id: str = None):
    """
    Fast single-call stock analysis.
    1. Get predictions (no LLM)
    2. Get news (no LLM)
    3. Single LLM call to generate report
    """
    ticker_upper = ticker.upper()
    query_vec = None
    mem = None

    # ---------------------------------------------------------
    # 1. CHECK SEMANTIC CACHE (Qdrant)
    # ---------------------------------------------------------
    embedder = get_embeddings_client()

    if embedder:
        try:
            mem = SemanticCache(collection_name="dataset_cache")
            query_text = f"Analysis report for {ticker_upper}"
            query_vec = embedder.embed_query(query_text)

            hits = mem.recall(query_vec, ticker=ticker_upper, limit=5)
            valid_hits = [h for h in hits if h.score > 0.95]

            if valid_hits:
                valid_hits.sort(key=lambda x: x.payload.get("created_at_ts", 0), reverse=True)
                cached_payload = valid_hits[0].payload

                if cached_payload.get("ticker") == ticker_upper:
                    print(f"✅ Semantic Cache HIT for {ticker_upper}", file=sys.stderr)
                    return {
                        "final_report": cached_payload.get("summary"),
                        "recommendation": cached_payload.get("recommendation", "Neutral"),
                        "confidence": cached_payload.get("confidence", "Medium"),
                        "last_price": cached_payload.get("last_price", 0.0),
                        "predictions": cached_payload.get("predictions", {})
                    }
        except Exception as e:
            print(f"⚠️ Semantic Cache Error: {e}", file=sys.stderr)

    # ---------------------------------------------------------
    # 2. FETCH PREDICTIONS
    # ---------------------------------------------------------
    logger.info(f"📊 Fetching predictions for {ticker_upper}")
    raw_data = fetch_prediction_data(ticker)

    if raw_data == "__MODEL_TRAINING__":
        return {
            "status": "training",
            "detail": f"Model for {ticker} is being trained. Retry after a few seconds.",
            "ticker": ticker_upper
        }

    if isinstance(raw_data, str):
        return {
            "status": "error",
            "detail": raw_data,
            "ticker": ticker_upper
        }

    if raw_data is None or not isinstance(raw_data, dict):
        return {
            "status": "error",
            "detail": f"Invalid prediction response for {ticker}: received {type(raw_data).__name__}",
            "ticker": ticker_upper
        }

    # Format predictions
    try:
        forecast = raw_data.get("result", {}).get("predictions", {}).get("full_forecast", [])
        history = raw_data.get("result", {}).get("history", [])

        if forecast:
            pred_lines = [f"5-Day Price Forecast for {ticker_upper}:"]
            for row in forecast[:5]:
                pred_lines.append(f"  {row['date']}: ${float(row.get('close', 0)):.2f}")
            pred_str = "\n".join(pred_lines)

            # Get last known price
            if history:
                last_price = float(history[-1].get("close", 0))
            else:
                last_price = float(forecast[0].get("close", 0))
        else:
            pred_str = f"No predictions available for {ticker_upper}"
            last_price = 0.0
    except Exception as e:
        pred_str = f"Prediction parsing failed: {e}"
        last_price = 0.0

    # ---------------------------------------------------------
    # 3. FETCH NEWS
    # ---------------------------------------------------------
    logger.info(f"📰 Fetching news for {ticker_upper}")
    try:
        news = get_stock_news(ticker_upper)
    except Exception as e:
        news = f"News unavailable: {e}"

    # ---------------------------------------------------------
    # 4. SINGLE LLM CALL FOR ANALYSIS
    # ---------------------------------------------------------
    logger.info(f"🤖 Generating analysis for {ticker_upper}")

    prompt = f"""Write a brief investment report for {ticker_upper} stock.

PRICE FORECAST:
{pred_str}

RECENT NEWS:
{news[:1000]}

Write 3-4 paragraphs covering price outlook and news sentiment. End with:
**Market Stance:** BULLISH or BEARISH or NEUTRAL
**Confidence:** High or Medium or Low

Do not use thinking tags. Write the report directly."""

    try:
        resp = invoke_llm_with_retry([HumanMessage(content=prompt)])
        report = resp.content if hasattr(resp, "content") else str(resp)
    except Exception as e:
        logger.error(f"LLM Error: {e}")
        # Return a basic report without LLM
        report = f"""## {ticker_upper} Analysis

### Price Forecast
{pred_str}

### News Summary
{news[:500]}

**Market Stance:** NEUTRAL
**Confidence:** Low

*Note: AI analysis unavailable. Showing raw data only.*"""

    # Extract recommendation
    upper_report = report.upper()
    if "BULLISH" in upper_report:
        recommendation = "BULLISH"
    elif "BEARISH" in upper_report:
        recommendation = "BEARISH"
    else:
        recommendation = "NEUTRAL"

    # Extract confidence
    if "HIGH" in upper_report:
        confidence = "High"
    elif "LOW" in upper_report:
        confidence = "Low"
    else:
        confidence = "Medium"

    result = {
        "ticker": ticker_upper,
        "final_report": report,
        "recommendation": recommendation,
        "confidence": confidence,
        "predictions": raw_data.get("result", {}).get("predictions", {}),
        "news_sentiment": news[:500]
    }

    # ---------------------------------------------------------
    # 5. SAVE TO CACHE
    # ---------------------------------------------------------
    if embedder and mem and query_vec:
        try:
            preds_data = result.get("predictions", {})
            mem.save_episode(
                ticker=ticker_upper,
                summary=report,
                embedding=query_vec,
                recommendation=recommendation,
                confidence=confidence,
                last_price=last_price,
                predictions=preds_data if isinstance(preds_data, dict) else {}
            )
            print(f"✅ Saved to Qdrant: {ticker_upper}", file=sys.stderr)
        except Exception as e:
            print(f"⚠️ Failed to save cache: {e}", file=sys.stderr)

    return result
