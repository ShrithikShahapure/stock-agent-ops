"""SageMaker Training script: train parent LSTM model on ^GSPC.

Runs inside a SageMaker PyTorch training container.
Reads preprocessed CSV from the input channel, trains the parent LSTM model,
and saves model artifacts to /opt/ml/model/ (uploaded to S3 by SageMaker).
"""

import argparse
import json
import os
import random
import sys

import joblib
import numpy as np
import pandas as pd
import torch
import torch.nn as nn
from sklearn.metrics import mean_squared_error, r2_score
from sklearn.preprocessing import StandardScaler
from torch.utils.data import DataLoader, Dataset

SEED = 42
FEATURES = ["Open", "High", "Low", "Close", "Volume", "RSI14", "MACD"]


def set_seeds(seed: int = SEED):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    if torch.cuda.is_available():
        torch.cuda.manual_seed_all(seed)
    torch.backends.cudnn.deterministic = True
    torch.backends.cudnn.benchmark = False


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


class StockDataset(Dataset):
    def __init__(self, df: pd.DataFrame, scaler: StandardScaler,
                 context_len: int = 10, pred_len: int = 1):
        vals = scaler.transform(df[FEATURES]).astype("float32")
        self.samples = []
        for t in range(context_len, len(df) - pred_len):
            past = vals[t - context_len:t]
            fut = vals[t:t + pred_len]
            if past.shape == (context_len, len(FEATURES)) and fut.shape == (pred_len, len(FEATURES)):
                self.samples.append((past, fut))
        if not self.samples:
            raise ValueError("No valid samples created")

    def __len__(self):
        return len(self.samples)

    def __getitem__(self, idx):
        past, fut = self.samples[idx]
        return torch.tensor(past), torch.tensor(fut)


def temporal_split(df, scaler, train_ratio=0.8, context_len=10, pred_len=1):
    split_idx = int(len(df) * train_ratio)
    train_df = df.iloc[:split_idx]
    val_start = max(0, split_idx - context_len)
    val_df = df.iloc[val_start:]
    train_ds = StockDataset(train_df, scaler, context_len, pred_len)
    val_ds = StockDataset(val_df, scaler, context_len, pred_len)
    return train_ds, val_ds


def fit_model(model, train_loader, val_loader, epochs, lr, device):
    model.to(device)
    opt = torch.optim.Adam(model.parameters(), lr=lr, weight_decay=1e-6)
    criterion = nn.MSELoss()
    scheduler = torch.optim.lr_scheduler.ReduceLROnPlateau(
        opt, mode="min", factor=0.5, patience=1)
    best_val_loss = float("inf")
    patience, counter = 2, 0

    for ep in range(1, epochs + 1):
        model.train()
        total_loss = 0.0
        for X, Y in train_loader:
            X, Y = X.to(device), Y.to(device)
            opt.zero_grad()
            pred = model(X)
            loss = criterion(pred, Y)
            loss.backward()
            torch.nn.utils.clip_grad_norm_(model.parameters(), 5.0)
            opt.step()
            total_loss += loss.item()
        avg_train = total_loss / len(train_loader)

        model.eval()
        val_loss = 0.0
        with torch.no_grad():
            for X, Y in val_loader:
                X, Y = X.to(device), Y.to(device)
                val_loss += criterion(model(X), Y).item()
        avg_val = val_loss / len(val_loader)

        print(f"Epoch {ep}/{epochs} - train_loss: {avg_train:.5f}, val_loss: {avg_val:.5f}")

        scheduler.step(avg_val)
        if avg_val < best_val_loss:
            best_val_loss = avg_val
            counter = 0
        else:
            counter += 1
            if counter >= patience:
                print("Early stopping triggered")
                break

    return model


def evaluate(model, df, scaler, context_len, pred_len, device):
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
    return {"MSE": mse, "RMSE": rmse, "R2": r2}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--epochs", type=int, default=3)
    parser.add_argument("--batch-size", type=int, default=128)
    parser.add_argument("--lr", type=float, default=1e-3)
    parser.add_argument("--context-len", type=int, default=10)
    parser.add_argument("--pred-len", type=int, default=1)
    parser.add_argument("--hidden-size", type=int, default=32)
    parser.add_argument("--parent-ticker", type=str, default="^GSPC")
    args = parser.parse_args()

    set_seeds()

    device = "cuda" if torch.cuda.is_available() else "cpu"
    print(f"Device: {device}")

    # SageMaker channels
    input_dir = os.environ.get("SM_CHANNEL_TRAINING", "/opt/ml/input/data/training")
    model_dir = os.environ.get("SM_MODEL_DIR", "/opt/ml/model")
    output_dir = os.environ.get("SM_OUTPUT_DATA_DIR", "/opt/ml/output/data")
    os.makedirs(model_dir, exist_ok=True)
    os.makedirs(output_dir, exist_ok=True)

    # Load parent data
    safe_ticker = args.parent_ticker.replace("^", "").replace("/", "_")
    csv_path = os.path.join(input_dir, f"{safe_ticker}.csv")
    print(f"Loading data from {csv_path}")
    df = pd.read_csv(csv_path)
    print(f"Loaded {len(df)} rows for {args.parent_ticker}")

    # Fit scaler and split
    scaler = StandardScaler().fit(df[FEATURES])
    train_ds, val_ds = temporal_split(
        df, scaler, train_ratio=0.8,
        context_len=args.context_len, pred_len=args.pred_len)
    train_loader = DataLoader(train_ds, batch_size=args.batch_size, shuffle=True)
    val_loader = DataLoader(val_ds, batch_size=args.batch_size)

    print(f"Training samples: {len(train_ds)}, Validation samples: {len(val_ds)}")

    # Train
    model = LSTMModel(
        input_size=len(FEATURES),
        hidden_size=args.hidden_size,
        pred_len=args.pred_len,
    )
    model = fit_model(model, train_loader, val_loader,
                      epochs=args.epochs, lr=args.lr, device=device)

    # Save model and scaler
    model_path = os.path.join(model_dir, f"{safe_ticker}_parent_model.pt")
    scaler_path = os.path.join(model_dir, f"{safe_ticker}_parent_scaler.pkl")
    torch.save(model.state_dict(), model_path)
    joblib.dump(scaler, scaler_path)
    print(f"Saved model to {model_path}")
    print(f"Saved scaler to {scaler_path}")

    # Evaluate on validation split
    split_idx = int(len(df) * 0.8)
    val_df = df.iloc[split_idx:]
    metrics = evaluate(model, val_df, scaler, args.context_len, args.pred_len, device)
    print(f"Validation metrics: {json.dumps(metrics, indent=2)}")

    # Save metrics for SageMaker
    metrics_path = os.path.join(output_dir, "metrics.json")
    with open(metrics_path, "w") as f:
        json.dump({"parent": {args.parent_ticker: metrics}}, f, indent=2)

    # Save hyperparameters for reference
    hparams = {
        "ticker": args.parent_ticker,
        "epochs": args.epochs,
        "batch_size": args.batch_size,
        "lr": args.lr,
        "context_len": args.context_len,
        "pred_len": args.pred_len,
        "hidden_size": args.hidden_size,
        "train_samples": len(train_ds),
        "val_samples": len(val_ds),
        "device": device,
    }
    hparams_path = os.path.join(model_dir, "hyperparameters.json")
    with open(hparams_path, "w") as f:
        json.dump(hparams, f, indent=2)


if __name__ == "__main__":
    main()
