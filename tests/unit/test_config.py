"""Unit tests for src.config.Config."""
import os
import pytest


def test_config_defaults():
    """Config should have reasonable defaults that work without a GPU."""
    from src.config import Config

    cfg = Config()

    assert cfg.context_len == 60, "context window should be 60 trading days"
    assert cfg.pred_len == 5, "prediction length should be 5 trading days"
    assert cfg.input_size == len(cfg.features)
    assert cfg.parent_ticker == "^GSPC"


def test_config_features_list():
    """Features must contain the OHLCV columns the model was trained on."""
    from src.config import Config

    cfg = Config()
    required = {"Open", "High", "Low", "Close", "Volume"}
    assert required.issubset(set(cfg.features)), (
        f"Config.features missing required OHLCV columns: {required - set(cfg.features)}"
    )


def test_config_device_is_string():
    """Device should be a plain string ('cpu' or 'cuda')."""
    from src.config import Config

    cfg = Config()
    assert isinstance(cfg.device, str)
    assert cfg.device in ("cpu", "cuda", "mps"), f"Unexpected device: {cfg.device}"


def test_config_workdir_is_string():
    from src.config import Config

    cfg = Config()
    assert isinstance(cfg.workdir, str)
    assert len(cfg.workdir) > 0
