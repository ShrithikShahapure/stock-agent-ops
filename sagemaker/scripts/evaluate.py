"""SageMaker Processing script: evaluate all trained models.

Runs inside a SageMaker Processing container. Loads the parent model
and all child models, evaluates them on their respective validation
splits, and produces an aggregate evaluation report.
"""

import argparse
import json
import os
import subprocess
import sys

# Install runtime dependencies not in the base SKLearn container
subprocess.check_call([
    sys.executable, "-m", "pip", "install", "-q",
    "torch>=2.5.0", "joblib>=1.4.0",
])

import joblib
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from sklearn.metrics import mean_squared_error, r2_score

FEATURES = ["Open", "High", "Low", "Close", "Volume", "RSI14", "MACD"]


class LSTMModel(nn.Module):
    def __init__(self, input_size: int = 7, hidden_size: int = 32,
                 num_layers: int = 1, pred_len: int = 1, dropout: float = 0.15):
        super().__init__()
        self.lstm = nn.LSTM(input_size, hidden_size, num_layers,
                            batch_first=True,
                            dropout=dropout if num_layers > 1 else 0)
        self.fc = nn.Linear(hidden_size, input_size * pred_len)
        self.pred_len = pred_len
        self.input_size = input_size

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        out, _ = self.lstm(x)
        out = out[:, -1, :]
        out = self.fc(out)
        return out.view(-1, self.pred_len, self.input_size)


def evaluate_model(model, df, scaler, context_len, pred_len, device):
    """Evaluate a model on a dataframe, returning MSE/RMSE/R2 on OHLCV."""
    vals = scaler.transform(df[FEATURES]).astype("float32")
    X_list, Y_list = [], []
    for t in range(context_len, len(vals) - pred_len):
        past = vals[t - context_len:t]
        fut = vals[t:t + pred_len]
        if past.shape == (context_len, len(FEATURES)) and fut.shape == (pred_len, len(FEATURES)):
            X_list.append(past)
            Y_list.append(fut)

    if not X_list:
        return {}

    model.eval()
    preds = []
    with torch.no_grad():
        for x in X_list:
            xt = torch.tensor(x.reshape(1, context_len, len(FEATURES)),
                              dtype=torch.float32).to(device)
            preds.append(model(xt).cpu().numpy()[0])

    Y_arr = np.array(Y_list).reshape(-1, len(FEATURES))[:, :5]
    P_arr = np.array(preds).reshape(-1, len(FEATURES))[:, :5]

    mse = float(mean_squared_error(Y_arr, P_arr))
    rmse = float(np.sqrt(mse))
    r2 = float(r2_score(Y_arr, P_arr))
    return {"MSE": mse, "RMSE": rmse, "R2": r2, "samples": len(X_list)}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--parent-ticker", type=str, default="^GSPC")
    parser.add_argument("--child-tickers", type=str, required=True)
    parser.add_argument("--context-len", type=int, default=10)
    parser.add_argument("--pred-len", type=int, default=1)
    parser.add_argument("--hidden-size", type=int, default=32)
    args = parser.parse_args()

    device = "cuda" if torch.cuda.is_available() else "cpu"

    # SageMaker processing paths
    data_dir = "/opt/ml/processing/input/data"
    parent_model_dir = "/opt/ml/processing/input/parent_model"
    child_model_dir = "/opt/ml/processing/input/child_model"
    output_dir = "/opt/ml/processing/output"
    os.makedirs(output_dir, exist_ok=True)

    report = {"parent": {}, "children": {}, "summary": {}}

    # Evaluate parent model
    safe_parent = args.parent_ticker.replace("^", "").replace("/", "_")
    parent_model_path = os.path.join(parent_model_dir, f"{safe_parent}_parent_model.pt")
    parent_scaler_path = os.path.join(parent_model_dir, f"{safe_parent}_parent_scaler.pkl")
    parent_csv = os.path.join(data_dir, f"{safe_parent}.csv")

    if all(os.path.exists(p) for p in [parent_model_path, parent_scaler_path, parent_csv]):
        print(f"Evaluating parent model: {args.parent_ticker}")
        model = LSTMModel(input_size=len(FEATURES), hidden_size=args.hidden_size,
                          pred_len=args.pred_len)
        model.load_state_dict(torch.load(parent_model_path, map_location=device,
                                         weights_only=True))
        scaler = joblib.load(parent_scaler_path)
        df = pd.read_csv(parent_csv)

        # Evaluate on validation split (last 20%)
        split_idx = int(len(df) * 0.8)
        val_df = df.iloc[split_idx:]
        metrics = evaluate_model(model, val_df, scaler,
                                 args.context_len, args.pred_len, device)
        report["parent"][args.parent_ticker] = metrics
        print(f"  Parent metrics: {json.dumps(metrics, indent=2)}")
    else:
        print(f"WARNING: Parent model artifacts not found", file=sys.stderr)

    # Evaluate child models
    child_tickers = [t.strip() for t in args.child_tickers.split(",")]
    all_r2 = []
    all_rmse = []

    for ticker in child_tickers:
        safe_name = ticker.replace("^", "").replace("/", "_").replace("-", "_")
        child_dir = os.path.join(child_model_dir, safe_name)
        model_path = os.path.join(child_dir, f"{safe_name}_child_model.pt")
        scaler_path = os.path.join(child_dir, f"{safe_name}_child_scaler.pkl")
        csv_path = os.path.join(data_dir, f"{safe_name}.csv")

        # Handle BRK-B -> BRK_B filename variants
        if not os.path.exists(csv_path):
            csv_path = os.path.join(data_dir, f"{ticker.replace('^', '').replace('-', '_')}.csv")

        if not all(os.path.exists(p) for p in [model_path, scaler_path, csv_path]):
            print(f"  Skipping {ticker}: artifacts not found")
            report["children"][ticker] = {"status": "skipped", "reason": "artifacts_missing"}
            continue

        try:
            print(f"Evaluating child model: {ticker}")
            model = LSTMModel(input_size=len(FEATURES), hidden_size=args.hidden_size,
                              pred_len=args.pred_len)
            model.load_state_dict(torch.load(model_path, map_location=device,
                                             weights_only=True))
            scaler = joblib.load(scaler_path)
            df = pd.read_csv(csv_path)

            split_idx = int(len(df) * 0.8)
            val_df = df.iloc[split_idx:]
            metrics = evaluate_model(model, val_df, scaler,
                                     args.context_len, args.pred_len, device)
            report["children"][ticker] = metrics
            if metrics.get("R2") is not None:
                all_r2.append(metrics["R2"])
                all_rmse.append(metrics["RMSE"])
            print(f"  {ticker} metrics: {json.dumps(metrics, indent=2)}")
        except Exception as e:
            print(f"  ERROR evaluating {ticker}: {e}", file=sys.stderr)
            report["children"][ticker] = {"status": "error", "error": str(e)}

    # Summary
    report["summary"] = {
        "total_children": len(child_tickers),
        "evaluated": len(all_r2),
        "mean_R2": float(np.mean(all_r2)) if all_r2 else None,
        "std_R2": float(np.std(all_r2)) if all_r2 else None,
        "mean_RMSE": float(np.mean(all_rmse)) if all_rmse else None,
        "std_RMSE": float(np.std(all_rmse)) if all_rmse else None,
        "best_ticker": child_tickers[int(np.argmax(all_r2))] if all_r2 else None,
        "worst_ticker": child_tickers[int(np.argmin(all_r2))] if all_r2 else None,
    }

    # Save evaluation report
    report_path = os.path.join(output_dir, "evaluation_report.json")
    with open(report_path, "w") as f:
        json.dump(report, f, indent=2)

    print(f"\nEvaluation complete. Report saved to {report_path}")
    print(f"Summary: {json.dumps(report['summary'], indent=2)}")


if __name__ == "__main__":
    main()
