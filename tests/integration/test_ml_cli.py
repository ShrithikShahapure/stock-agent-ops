"""
Integration tests for scripts/ml_cli.py.

These tests exercise the CLI argument parsing and command dispatch
using mocked pipeline functions — no real model training or network
calls are made.
"""
import json
import subprocess
import sys
from pathlib import Path
from unittest.mock import patch

import pytest

ROOT = Path(__file__).resolve().parents[2]
CLI = ROOT / "scripts" / "ml_cli.py"


def run_cli(*args, env=None) -> subprocess.CompletedProcess:
    """Run ml_cli.py in a subprocess and return the completed process."""
    import os
    e = os.environ.copy()
    if env:
        e.update(env)
    return subprocess.run(
        [sys.executable, str(CLI), *args],
        capture_output=True,
        text=True,
        env=e,
        cwd=str(ROOT),
    )


# ---------------------------------------------------------------------------
# CLI structure / help
# ---------------------------------------------------------------------------

class TestCLIHelp:
    def test_no_args_exits_nonzero(self):
        proc = run_cli()
        assert proc.returncode != 0, "CLI with no args should exit non-zero"

    def test_help_flag(self):
        proc = run_cli("--help")
        # argparse help exits 0
        assert proc.returncode == 0
        assert "usage" in proc.stdout.lower() or "usage" in proc.stderr.lower()

    def test_unknown_command_exits_nonzero(self):
        proc = run_cli("not-a-command")
        assert proc.returncode != 0


# ---------------------------------------------------------------------------
# Command: predict-child
# ---------------------------------------------------------------------------

class TestPredictChildCLI:
    def test_missing_ticker_exits_nonzero(self):
        proc = run_cli("predict-child")
        assert proc.returncode != 0

    def test_with_ticker_calls_pipeline(self):
        """predict-child --ticker AAPL should call predict_child and print JSON."""
        mock_result = {
            "ticker": "AAPL",
            "predictions": {"full_forecast": []},
        }
        with patch("src.pipelines.inference_pipeline.predict_child", return_value=mock_result):
            import importlib
            import scripts.ml_cli as cli_mod
            importlib.reload(cli_mod)

            # Parse and dispatch directly (avoids subprocess overhead for mock)
            import io, contextlib
            buf = io.StringIO()
            with contextlib.redirect_stdout(buf):
                cli_mod.predict_child("AAPL")

            output = buf.getvalue()
            data = json.loads(output)
            assert data["ticker"] == "AAPL"


# ---------------------------------------------------------------------------
# Command: predict-parent
# ---------------------------------------------------------------------------

class TestPredictParentCLI:
    def test_calls_pipeline_and_outputs_json(self):
        mock_result = {"ticker": "^GSPC", "predictions": {}}
        with patch("src.pipelines.inference_pipeline.predict_parent", return_value=mock_result):
            import importlib
            import scripts.ml_cli as cli_mod
            importlib.reload(cli_mod)

            import io, contextlib
            buf = io.StringIO()
            with contextlib.redirect_stdout(buf):
                cli_mod.predict_parent()

            data = json.loads(buf.getvalue())
            assert data["ticker"] == "^GSPC"


# ---------------------------------------------------------------------------
# Command: analyze
# ---------------------------------------------------------------------------

class TestAnalyzeCLI:
    def test_missing_ticker_exits_nonzero(self):
        proc = run_cli("analyze")
        assert proc.returncode != 0

    def test_outputs_json_with_ticker(self):
        mock_result = {
            "ticker": "NVDA",
            "final_report": "Some report",
            "recommendation": "BULLISH",
            "confidence": "High",
        }
        with patch("src.agents.graph.analyze_stock", return_value=mock_result):
            import importlib
            import scripts.ml_cli as cli_mod
            importlib.reload(cli_mod)

            import io, contextlib
            buf = io.StringIO()
            with contextlib.redirect_stdout(buf):
                cli_mod.analyze("NVDA", thread_id="test-123")

            data = json.loads(buf.getvalue())
            assert data["ticker"] == "NVDA"
            assert data["recommendation"] == "BULLISH"


# ---------------------------------------------------------------------------
# Error handling: pipeline exceptions become JSON error + exit 1
# ---------------------------------------------------------------------------

class TestCLIErrorHandling:
    def test_predict_child_exception_outputs_json_error(self):
        """If the pipeline raises, CLI should print JSON {error: ...} and exit 1."""
        from src.exception import PipelineError

        with patch(
            "src.pipelines.inference_pipeline.predict_child",
            side_effect=PipelineError("Model file missing"),
        ):
            import importlib
            import scripts.ml_cli as cli_mod
            importlib.reload(cli_mod)

            import io, contextlib
            buf = io.StringIO()
            with pytest.raises(SystemExit) as exc_info:
                with contextlib.redirect_stdout(buf):
                    cli_mod.predict_child("AAPL")

            assert exc_info.value.code == 1
            data = json.loads(buf.getvalue())
            assert "error" in data
