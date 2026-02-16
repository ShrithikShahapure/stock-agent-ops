#!/usr/bin/env python3
"""
ML CLI Wrapper for Stock Agent Ops

This script provides a command-line interface for the ML pipelines and agents.
It outputs JSON to stdout and returns exit code 0 on success, 1 on error.

Usage:
    python scripts/ml_cli.py train-parent
    python scripts/ml_cli.py train-child --ticker AAPL
    python scripts/ml_cli.py predict-parent
    python scripts/ml_cli.py predict-child --ticker AAPL
    python scripts/ml_cli.py analyze --ticker AAPL [--thread-id ID]
    python scripts/ml_cli.py monitor-parent
    python scripts/ml_cli.py monitor-ticker --ticker AAPL
"""

import sys
import os
import json
import argparse
import traceback

# Add the project root to the path
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)


def output_json(data):
    """Print JSON to stdout."""
    print(json.dumps(data, default=str))


def output_error(message, details=None):
    """Print error JSON to stdout and exit with code 1."""
    error_data = {"error": message}
    if details:
        error_data["details"] = details
    output_json(error_data)
    sys.exit(1)


def train_parent():
    """Train the parent model (S&P 500)."""
    try:
        from src.pipelines.training_pipeline import train_parent as _train_parent
        result = _train_parent()
        return {"status": "completed", "result": result}
    except Exception as e:
        output_error(f"Training failed: {e}", traceback.format_exc())


def train_child(ticker):
    """Train a child model for a specific ticker."""
    try:
        from src.pipelines.training_pipeline import train_child as _train_child
        result = _train_child(ticker.upper())
        return {"status": "completed", "ticker": ticker.upper(), "result": result}
    except Exception as e:
        output_error(f"Training failed for {ticker}: {e}", traceback.format_exc())


def predict_parent():
    """Get predictions from the parent model."""
    try:
        from src.pipelines.inference_pipeline import predict_parent as _predict_parent
        result = _predict_parent()
        return result
    except FileNotFoundError as e:
        output_error(f"Missing model: {e}")
    except Exception as e:
        output_error(f"Prediction failed: {e}", traceback.format_exc())


def predict_child(ticker):
    """Get predictions from a child model."""
    try:
        from src.pipelines.inference_pipeline import predict_child as _predict_child
        result = _predict_child(ticker.upper())
        return result
    except FileNotFoundError as e:
        output_error(f"Missing model for {ticker}: {e}")
    except Exception as e:
        # Check if it's a model missing error
        error_str = str(e).lower()
        if "missing" in error_str or "not found" in error_str:
            output_error(f"Missing model for {ticker}: {e}")
        output_error(f"Prediction failed for {ticker}: {e}", traceback.format_exc())


def analyze(ticker, thread_id=None):
    """Run AI agent analysis on a ticker."""
    try:
        from src.agents.graph import analyze_stock
        result = analyze_stock(ticker.upper(), thread_id=thread_id)
        return result
    except Exception as e:
        output_error(f"Analysis failed for {ticker}: {e}", traceback.format_exc())


def monitor_parent():
    """Run monitoring on the parent model."""
    try:
        from src.config import Config
        from src.monitoring.drift import check_drift
        from src.monitoring.agent_eval import AgentEvaluator

        cfg = Config()
        ticker = cfg.parent_ticker
        base_path = "outputs"

        # Run drift check
        try:
            drift_res = check_drift(ticker, base_path)
        except Exception as e:
            drift_res = {"status": "failed", "error": str(e)}

        # Run agent evaluation
        try:
            evaluator = AgentEvaluator(base_path)
            eval_res = evaluator.evaluate_live(ticker)
        except Exception as e:
            eval_res = {"status": "failed", "error": str(e)}

        return {
            "ticker": ticker,
            "type": "Parent Model (Market Index)",
            "drift": drift_res,
            "agent_eval": eval_res
        }
    except Exception as e:
        output_error(f"Monitoring failed: {e}", traceback.format_exc())


def monitor_ticker(ticker):
    """Run monitoring on a specific ticker."""
    try:
        from src.config import Config
        from src.monitoring.drift import check_drift
        from src.monitoring.agent_eval import AgentEvaluator

        cfg = Config()
        clean_ticker = ticker.strip().upper()
        is_parent = (clean_ticker == cfg.parent_ticker)
        base_path = "outputs"

        # Run drift check only for parent
        if is_parent:
            try:
                drift_res = check_drift(clean_ticker, base_path)
            except Exception as e:
                drift_res = {"status": "failed", "error": str(e)}
        else:
            drift_res = {"status": "skipped", "detail": "Drift calculation reserved for parent model."}

        # Run agent evaluation
        try:
            evaluator = AgentEvaluator(base_path)
            eval_res = evaluator.evaluate_live(clean_ticker)
        except Exception as e:
            eval_res = {"status": "failed", "error": str(e)}

        return {
            "ticker": clean_ticker,
            "is_parent": is_parent,
            "drift": drift_res,
            "agent_eval": eval_res
        }
    except Exception as e:
        output_error(f"Monitoring failed for {ticker}: {e}", traceback.format_exc())


def main():
    parser = argparse.ArgumentParser(
        description="ML CLI Wrapper for Stock Agent Ops",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    # train-parent
    subparsers.add_parser("train-parent", help="Train the parent model (S&P 500)")

    # train-child
    tc = subparsers.add_parser("train-child", help="Train a child model for a specific ticker")
    tc.add_argument("--ticker", required=True, help="Stock ticker symbol (e.g., AAPL)")

    # predict-parent
    subparsers.add_parser("predict-parent", help="Get predictions from the parent model")

    # predict-child
    pc = subparsers.add_parser("predict-child", help="Get predictions from a child model")
    pc.add_argument("--ticker", required=True, help="Stock ticker symbol (e.g., AAPL)")

    # analyze
    az = subparsers.add_parser("analyze", help="Run AI agent analysis on a ticker")
    az.add_argument("--ticker", required=True, help="Stock ticker symbol (e.g., AAPL)")
    az.add_argument("--thread-id", default=None, help="Optional thread ID for conversation memory")

    # monitor-parent
    subparsers.add_parser("monitor-parent", help="Run monitoring on the parent model")

    # monitor-ticker
    mt = subparsers.add_parser("monitor-ticker", help="Run monitoring on a specific ticker")
    mt.add_argument("--ticker", required=True, help="Stock ticker symbol (e.g., AAPL)")

    args = parser.parse_args()

    try:
        if args.command == "train-parent":
            result = train_parent()
        elif args.command == "train-child":
            result = train_child(args.ticker)
        elif args.command == "predict-parent":
            result = predict_parent()
        elif args.command == "predict-child":
            result = predict_child(args.ticker)
        elif args.command == "analyze":
            result = analyze(args.ticker, args.thread_id)
        elif args.command == "monitor-parent":
            result = monitor_parent()
        elif args.command == "monitor-ticker":
            result = monitor_ticker(args.ticker)
        else:
            output_error(f"Unknown command: {args.command}")

        output_json(result)
        sys.exit(0)

    except SystemExit:
        raise
    except Exception as e:
        output_error(f"Unexpected error: {e}", traceback.format_exc())


if __name__ == "__main__":
    main()
