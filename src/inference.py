import math
import numpy as np
import pandas as pd
from typing import Dict
# import onnxruntime as ort
from sklearn.preprocessing import StandardScaler
from src.config import Config
from logger.logger import get_logger
from src.exception import PipelineError

logger = get_logger()

# Trading-day counts for each horizon
HORIZONS = {"week": 5, "month": 21, "quarter": 63}


def _roll_forecast(model, vals: np.ndarray, cfg: "Config", total_days: int) -> np.ndarray:
    """
    Autoregressively roll the model forward to produce `total_days` scaled predictions.
    The model predicts cfg.pred_len steps per call; we chain calls until we have enough.
    """
    import torch

    window = vals[-cfg.context_len:].copy()  # shape (context_len, input_size)
    all_preds: list[np.ndarray] = []

    rounds = math.ceil(total_days / cfg.pred_len)
    for _ in range(rounds):
        x = window[-cfg.context_len:].reshape(1, cfg.context_len, cfg.input_size)
        x_tensor = torch.tensor(x, dtype=torch.float32).to(cfg.device)
        with torch.no_grad():
            pred = model(x_tensor).cpu().numpy()[0]  # (pred_len, input_size)
        all_preds.append(pred)
        window = np.vstack([window, pred])

    return np.vstack(all_preds)[:total_days]  # (total_days, input_size)


def _build_forecast_list(preds_inv: np.ndarray, last_date: pd.Timestamp) -> list[dict]:
    """Convert inverse-scaled OHLCV array into a list of dated dicts."""
    days = pd.bdate_range(last_date + pd.Timedelta(days=1), periods=len(preds_inv))
    return [
        {
            "date": str(d.date()),
            "open": float(preds_inv[i][0]),
            "high": float(preds_inv[i][1]),
            "low": float(preds_inv[i][2]),
            "close": float(preds_inv[i][3]),
            "volume": float(preds_inv[i][4]),
        }
        for i, d in enumerate(days)
    ]


def predict_multi_horizon(model, df: pd.DataFrame, scaler: StandardScaler, ticker: str) -> Dict:
    """
    Generate multi-horizon forecasts using autoregressive rolling inference.

    The model's pred_len (5) stays unchanged — no retraining needed. We chain
    model calls to produce predictions for:
      week    →  5 trading days
      month   → 21 trading days
      quarter → 63 trading days

    Returns a unified dict with all three horizons and the full 63-day array
    under `full_forecast` for backward-compatible chart rendering.
    """
    try:
        cfg = Config()
        vals = scaler.transform(df[cfg.features]).astype("float32")
        last_date = df["date"].iloc[-1]

        total_days = HORIZONS["quarter"]  # 63 — longest horizon
        scaled_preds = _roll_forecast(model, vals, cfg, total_days)  # (63, input_size)
        preds_inv = scaler.inverse_transform(scaled_preds)[:, :5]    # (63, 5) OHLCV

        full = _build_forecast_list(preds_inv, last_date)
        week_fc    = full[:HORIZONS["week"]]    #  5 days
        month_fc   = full[:HORIZONS["month"]]   # 21 days
        quarter_fc = full[:HORIZONS["quarter"]] # 63 days

        return {
            "ticker": ticker,
            "last_date": str(last_date.date()),
            "future_window_days": total_days,
            "next_business_days": [d["date"] for d in week_fc],
            "predictions": {
                "next_day": full[0],
                "week": week_fc,
                "month": month_fc,
                "quarter": quarter_fc,
                "full_forecast": quarter_fc,  # alias kept for backward compat
            },
        }

    except Exception as e:
        logger.error(f"Prediction failed for {ticker}: {e}")
        raise PipelineError(f"Prediction failed for {ticker}: {e}")


# Backward-compatible alias
predict_one_step_and_week = predict_multi_horizon
