"""Unit tests for src.agents.graph.analyze_stock."""
from unittest.mock import MagicMock, patch

import pytest


# Shared prediction response that mimics what predict-child returns
MOCK_PREDICT_RESPONSE = {
    "result": {
        "predictions": {
            "full_forecast": [
                {"date": "2025-01-06", "close": 150.0},
                {"date": "2025-01-07", "close": 151.0},
                {"date": "2025-01-08", "close": 152.0},
                {"date": "2025-01-09", "close": 153.0},
                {"date": "2025-01-10", "close": 154.0},
            ],
            "week": [],
            "month": [],
            "quarter": [],
        },
        "history": [
            {"date": "2024-12-31", "close": 148.5},
        ],
    }
}


def _patch_analyze(
    predict_data=None,
    embedder=None,
    news="Some news headline",
    llm_response="AAPL looks good.\n**Market Stance:** BULLISH\n**Confidence:** High",
):
    """Context manager that patches all external calls in analyze_stock."""
    from langchain_core.messages import AIMessage

    predict_data = predict_data if predict_data is not None else MOCK_PREDICT_RESPONSE

    return (
        patch("src.agents.graph.get_embeddings_client", return_value=embedder),
        patch("src.agents.graph.fetch_prediction_data", return_value=predict_data),
        patch("src.agents.graph.get_stock_news", return_value=news),
        patch(
            "src.agents.graph.invoke_llm_with_retry",
            return_value=AIMessage(content=llm_response),
        ),
    )


class TestAnalyzeStockSuccess:
    def test_returns_required_keys(self):
        from src.agents.graph import analyze_stock

        patches = _patch_analyze()
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        for key in ("ticker", "final_report", "recommendation", "confidence", "predictions"):
            assert key in result, f"Missing key '{key}' in analyze_stock result"

    def test_ticker_is_uppercased(self):
        from src.agents.graph import analyze_stock

        patches = _patch_analyze()
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("aapl")

        assert result["ticker"] == "AAPL"

    def test_bullish_recommendation_detected(self):
        from src.agents.graph import analyze_stock

        patches = _patch_analyze(
            llm_response="Strong performance.\n**Market Stance:** BULLISH\n**Confidence:** High"
        )
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        assert result["recommendation"] == "BULLISH"
        assert result["confidence"] == "High"

    def test_bearish_recommendation_detected(self):
        from src.agents.graph import analyze_stock

        patches = _patch_analyze(
            llm_response="Declining trend.\n**Market Stance:** BEARISH\n**Confidence:** Medium"
        )
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        assert result["recommendation"] == "BEARISH"

    def test_neutral_recommendation_is_default(self):
        from src.agents.graph import analyze_stock

        patches = _patch_analyze(llm_response="Uncertain outlook. No clear direction.")
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        assert result["recommendation"] == "NEUTRAL"

    def test_predictions_includes_history(self):
        from src.agents.graph import analyze_stock

        patches = _patch_analyze()
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        assert "history" in result["predictions"], (
            "predictions dict must include history key for frontend chart"
        )


class TestAnalyzeStockErrorCases:
    def test_model_training_state_propagated(self):
        from src.agents.graph import analyze_stock

        patches = (
            patch("src.agents.graph.get_embeddings_client", return_value=None),
            patch("src.agents.graph.fetch_prediction_data", return_value="__MODEL_TRAINING__"),
            patch("src.agents.graph.get_stock_news", return_value=""),
            patch("src.agents.graph.invoke_llm_with_retry", return_value=MagicMock()),
        )
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        assert result["status"] == "training"
        assert result["ticker"] == "AAPL"

    def test_invalid_prediction_response_returns_error(self):
        from src.agents.graph import analyze_stock

        patches = (
            patch("src.agents.graph.get_embeddings_client", return_value=None),
            patch("src.agents.graph.fetch_prediction_data", return_value="some error string"),
            patch("src.agents.graph.get_stock_news", return_value=""),
            patch("src.agents.graph.invoke_llm_with_retry", return_value=MagicMock()),
        )
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        assert result["status"] == "error"

    def test_llm_failure_returns_fallback_report(self):
        from src.agents.graph import analyze_stock

        patches = (
            patch("src.agents.graph.get_embeddings_client", return_value=None),
            patch("src.agents.graph.fetch_prediction_data", return_value=MOCK_PREDICT_RESPONSE),
            patch("src.agents.graph.get_stock_news", return_value="News data"),
            patch("src.agents.graph.invoke_llm_with_retry", side_effect=Exception("LLM timeout")),
        )
        with patches[0], patches[1], patches[2], patches[3]:
            result = analyze_stock("AAPL")

        # Should still return a usable result (fallback report)
        assert "final_report" in result
        assert result["ticker"] == "AAPL"


class TestAnalyzeStockSemanticCache:
    def test_cache_hit_skips_llm(self):
        """When Qdrant returns a high-score hit, the LLM should not be called."""
        from src.agents.graph import analyze_stock

        cached_payload = {
            "ticker": "AAPL",
            "summary": "Cached report",
            "recommendation": "BULLISH",
            "confidence": "High",
            "last_price": 150.0,
            "predictions": {},
            "created_at_ts": 9999999999,
        }

        mock_hit = MagicMock()
        mock_hit.score = 0.98
        mock_hit.payload = cached_payload

        mock_cache = MagicMock()
        mock_cache.recall.return_value = [mock_hit]

        mock_embedder = MagicMock()
        mock_embedder.embed_query.return_value = [0.1] * 768

        with (
            patch("src.agents.graph.get_embeddings_client", return_value=mock_embedder),
            patch("src.agents.graph.SemanticCache", return_value=mock_cache),
            patch("src.agents.graph.invoke_llm_with_retry") as mock_llm,
        ):
            result = analyze_stock("AAPL")

        mock_llm.assert_not_called()
        assert result["final_report"] == "Cached report"
        assert result["recommendation"] == "BULLISH"
