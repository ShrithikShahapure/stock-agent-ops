"""SageMaker Processing script: fetch OHLCV data for all tickers.

Runs inside a SageMaker Processing container. Fetches 1 year of daily
OHLCV data from yfinance for the parent ticker (^GSPC) and all child
tickers, computes RSI14 and MACD features, and saves CSV files to the
output directory (which SageMaker uploads to S3).
"""

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime, timedelta

# Install runtime dependencies not in the base SKLearn container
subprocess.check_call([sys.executable, "-m", "pip", "install", "-q", "yfinance>=0.2.40"])

import pandas as pd
import yfinance as yf


def rsi(series: pd.Series, period: int = 14) -> pd.Series:
    delta = series.diff()
    gain = delta.where(delta > 0, 0).rolling(window=period).mean()
    loss = -delta.where(delta < 0, 0).rolling(window=period).mean()
    rs = gain / loss
    return 100 - 100 / (1 + rs)


def macd(series: pd.Series, fast: int = 12, slow: int = 26) -> pd.Series:
    ema_fast = series.ewm(span=fast, adjust=False).mean()
    ema_slow = series.ewm(span=slow, adjust=False).mean()
    return ema_fast - ema_slow


FEATURES = ["Open", "High", "Low", "Close", "Volume", "RSI14", "MACD"]


def fetch_ticker(ticker: str, start_date: str, end_date: str) -> pd.DataFrame:
    """Fetch OHLCV data with technical indicators for a single ticker."""
    print(f"Fetching {ticker} from {start_date} to {end_date}")
    df = yf.download(ticker, start=start_date, end=end_date, interval="1d",
                     auto_adjust=True, progress=False)
    if df.empty:
        raise ValueError(f"No data downloaded for {ticker}")

    if isinstance(df.columns, pd.MultiIndex):
        df.columns = df.columns.get_level_values(0)

    df = df.reset_index().rename(columns={"Date": "date"})
    df = df[["date", "Open", "High", "Low", "Close", "Volume"]].dropna()
    df["RSI14"] = rsi(df["Close"])
    df["MACD"] = macd(df["Close"])
    df = df[["date"] + FEATURES].dropna()

    min_rows = 10 + 1  # context_len + pred_len
    if len(df) < min_rows:
        raise ValueError(f"Insufficient data for {ticker}: {len(df)} rows, need {min_rows}")

    print(f"  {ticker}: {len(df)} rows fetched")
    return df


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--tickers", type=str, required=True,
                        help="Comma-separated list of tickers (first is parent)")
    parser.add_argument("--lookback-days", type=int, default=365)
    parser.add_argument("--output-dir", type=str,
                        default="/opt/ml/processing/output")
    args = parser.parse_args()

    tickers = [t.strip() for t in args.tickers.split(",")]
    end_date = datetime.now().strftime("%Y-%m-%d")
    start_date = (datetime.now() - timedelta(days=args.lookback_days)).strftime("%Y-%m-%d")

    output_dir = args.output_dir
    os.makedirs(output_dir, exist_ok=True)

    manifest = {"start_date": start_date, "end_date": end_date, "tickers": {}}

    failed = []
    for ticker in tickers:
        try:
            df = fetch_ticker(ticker, start_date, end_date)
            # Sanitize ticker for filename (^GSPC -> GSPC)
            safe_name = ticker.replace("^", "").replace("/", "_")
            csv_path = os.path.join(output_dir, f"{safe_name}.csv")
            df.to_csv(csv_path, index=False)
            manifest["tickers"][ticker] = {
                "file": f"{safe_name}.csv",
                "rows": len(df),
                "start": str(df["date"].iloc[0]),
                "end": str(df["date"].iloc[-1]),
            }
        except Exception as e:
            print(f"ERROR fetching {ticker}: {e}", file=sys.stderr)
            failed.append(ticker)

    manifest["failed"] = failed
    manifest_path = os.path.join(output_dir, "manifest.json")
    with open(manifest_path, "w") as f:
        json.dump(manifest, f, indent=2)

    print(f"\nPreprocessing complete: {len(tickers) - len(failed)}/{len(tickers)} tickers fetched")
    if failed:
        print(f"Failed tickers: {failed}", file=sys.stderr)
        # Don't fail the job for individual ticker failures as long as parent succeeded
        if tickers[0] in failed:
            sys.exit(1)


if __name__ == "__main__":
    main()
