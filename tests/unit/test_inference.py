"""Unit tests for src.inference — multi-horizon rolling forecast logic."""
import math
from datetime import date, timedelta
from unittest.mock import MagicMock, patch

import numpy as np
import pandas as pd
import pytest


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_df(n_rows: int = 100) -> pd.DataFrame:
    """Return a minimal OHLCV DataFrame with a 'date' column."""
    dates = pd.bdate_range(end=pd.Timestamp("2025-01-01"), periods=n_rows)
    df = pd.DataFrame(
        {
            "date": dates,
            "Open": np.random.uniform(100, 200, n_rows),
            "High": np.random.uniform(200, 300, n_rows),
            "Low": np.random.uniform(50, 100, n_rows),
            "Close": np.random.uniform(100, 200, n_rows),
            "Volume": np.random.uniform(1e6, 1e7, n_rows),
            "RSI14": np.random.uniform(30, 70, n_rows),
            "MACD": np.random.uniform(-5, 5, n_rows),
        }
    )
    return df


def _make_mock_model(context_len: int = 60, pred_len: int = 5, input_size: int = 7):
    """Return a callable mock model that outputs ones of the right shape."""
    import torch

    def forward(x):
        batch = x.shape[0]
        return torch.ones(batch, pred_len, input_size)

    mock = MagicMock(side_effect=forward)
    mock.eval = MagicMock(return_value=None)
    return mock


# ---------------------------------------------------------------------------
# _build_forecast_list
# ---------------------------------------------------------------------------

class TestBuildForecastList:
    def test_length_matches_input(self):
        from src.inference import _build_forecast_list

        n = 10
        preds = np.random.rand(n, 5)
        last_date = pd.Timestamp("2025-01-01")
        result = _build_forecast_list(preds, last_date)
        assert len(result) == n

    def test_dates_are_business_days(self):
        from src.inference import _build_forecast_list

        preds = np.ones((5, 5))
        last_date = pd.Timestamp("2025-01-03")  # a Friday
        result = _build_forecast_list(preds, last_date)

        prev = last_date
        for item in result:
            d = pd.Timestamp(item["date"])
            assert d > prev, "Each forecast date must be after the previous"
            assert d.weekday() < 5, f"Forecast date {d} falls on a weekend"
            prev = d

    def test_ohlcv_keys_present(self):
        from src.inference import _build_forecast_list

        preds = np.ones((3, 5))
        result = _build_forecast_list(preds, pd.Timestamp("2025-01-01"))

        for item in result:
            for key in ("date", "open", "high", "low", "close", "volume"):
                assert key in item, f"Missing key '{key}' in forecast item"

    def test_values_are_floats(self):
        from src.inference import _build_forecast_list

        preds = np.array([[1.1, 2.2, 3.3, 4.4, 5.5]])
        result = _build_forecast_list(preds, pd.Timestamp("2025-01-01"))
        item = result[0]
        assert isinstance(item["close"], float)
        assert abs(item["close"] - 4.4) < 1e-6


# ---------------------------------------------------------------------------
# _roll_forecast
# ---------------------------------------------------------------------------

class TestRollForecast:
    def test_output_shape(self):
        import torch
        from src.inference import _roll_forecast

        cfg_mock = MagicMock()
        cfg_mock.context_len = 60
        cfg_mock.pred_len = 5
        cfg_mock.input_size = 7
        cfg_mock.device = "cpu"

        vals = np.random.rand(100, 7).astype("float32")

        # Simple model: returns ones of shape (pred_len, input_size)
        def model(x):
            batch = x.shape[0]
            return torch.ones(batch, cfg_mock.pred_len, cfg_mock.input_size)

        total = 63
        out = _roll_forecast(model, vals, cfg_mock, total)
        assert out.shape == (total, 7), f"Expected ({total}, 7), got {out.shape}"

    def test_single_step_equals_model_output(self):
        """When total_days == pred_len, the output should match one model call."""
        import torch
        from src.inference import _roll_forecast

        cfg_mock = MagicMock()
        cfg_mock.context_len = 10
        cfg_mock.pred_len = 5
        cfg_mock.input_size = 3
        cfg_mock.device = "cpu"

        expected = np.full((5, 3), 0.42, dtype="float32")

        def model(x):
            return torch.tensor(expected).unsqueeze(0)  # (1, 5, 3)

        vals = np.zeros((20, 3), dtype="float32")
        out = _roll_forecast(model, vals, cfg_mock, 5)
        np.testing.assert_allclose(out, expected, atol=1e-5)


# ---------------------------------------------------------------------------
# predict_multi_horizon (integration of both helpers)
# ---------------------------------------------------------------------------

class TestPredictMultiHorizon:
    def test_returns_all_horizon_keys(self):
        from src.inference import predict_multi_horizon
        from sklearn.preprocessing import StandardScaler

        df = _make_df()
        features = ["Open", "High", "Low", "Close", "Volume", "RSI14", "MACD"]

        scaler = StandardScaler()
        scaler.fit(df[features].values)

        cfg_mock = MagicMock()
        cfg_mock.context_len = 60
        cfg_mock.pred_len = 5
        cfg_mock.input_size = 7
        cfg_mock.features = features
        cfg_mock.device = "cpu"

        import torch

        def model(x):
            return torch.ones(x.shape[0], cfg_mock.pred_len, cfg_mock.input_size)

        with patch("src.inference.Config", return_value=cfg_mock):
            result = predict_multi_horizon(model, df, scaler, "AAPL")

        assert "ticker" in result
        assert "predictions" in result
        preds = result["predictions"]
        for key in ("next_day", "week", "month", "quarter", "full_forecast"):
            assert key in preds, f"Missing key '{key}' in predictions"

    def test_horizon_lengths(self):
        """week=5, month=21, quarter=63 items exactly."""
        from src.inference import predict_multi_horizon
        from sklearn.preprocessing import StandardScaler

        df = _make_df()
        features = ["Open", "High", "Low", "Close", "Volume", "RSI14", "MACD"]

        scaler = StandardScaler()
        scaler.fit(df[features].values)

        cfg_mock = MagicMock()
        cfg_mock.context_len = 60
        cfg_mock.pred_len = 5
        cfg_mock.input_size = 7
        cfg_mock.features = features
        cfg_mock.device = "cpu"

        import torch

        def model(x):
            return torch.ones(x.shape[0], cfg_mock.pred_len, cfg_mock.input_size)

        with patch("src.inference.Config", return_value=cfg_mock):
            result = predict_multi_horizon(model, df, scaler, "AAPL")

        preds = result["predictions"]
        assert len(preds["week"]) == 5
        assert len(preds["month"]) == 21
        assert len(preds["quarter"]) == 63
        assert len(preds["full_forecast"]) == 63

    def test_backward_compat_alias(self):
        """predict_one_step_and_week should be the same callable."""
        from src.inference import predict_multi_horizon, predict_one_step_and_week

        assert predict_one_step_and_week is predict_multi_horizon
