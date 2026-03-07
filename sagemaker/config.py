"""SageMaker pipeline configuration."""

import os
from dataclasses import dataclass, field
from typing import List


@dataclass
class SageMakerConfig:
    """Configuration for the SageMaker training pipeline."""

    # AWS
    region: str = os.environ.get("AWS_REGION", "us-east-1")
    role_arn: str = os.environ.get("SAGEMAKER_ROLE_ARN", "")
    bucket: str = os.environ.get("SAGEMAKER_BUCKET", "")
    pipeline_name: str = "stock-agent-ops-training"

    # S3 prefixes
    s3_prefix: str = "sagemaker"
    data_prefix: str = "data"
    model_prefix: str = "models"
    evaluation_prefix: str = "evaluation"

    # Instance types (C family on-demand)
    processing_instance_type: str = "ml.c5.xlarge"
    training_instance_type: str = "ml.c5.xlarge"
    processing_instance_count: int = 1
    training_instance_count: int = 1

    # Training volume
    training_volume_size_gb: int = 30

    # Timeouts (seconds)
    processing_timeout: int = 3600  # 1 hour
    training_timeout: int = 7200  # 2 hours

    # Model parameters
    parent_ticker: str = "^GSPC"
    parent_epochs: int = 3
    child_epochs: int = 2
    batch_size: int = 128
    context_len: int = 10
    pred_len: int = 1
    hidden_size: int = 32
    transfer_strategy: str = "freeze"

    # Top 20 S&P 500 stocks by market cap
    child_tickers: List[str] = field(default_factory=lambda: [
        "AAPL", "MSFT", "NVDA", "AMZN", "GOOG",
        "META", "BRK-B", "LLY", "AVGO", "JPM",
        "TSLA", "UNH", "V", "XOM", "MA",
        "PG", "COST", "JNJ", "HD", "WMT",
    ])

    # Data range: 1 year lookback
    lookback_days: int = 365

    # Features
    features: List[str] = field(default_factory=lambda: [
        "Open", "High", "Low", "Close", "Volume", "RSI14", "MACD",
    ])

    # Model registry
    model_package_group: str = "stock-agent-ops-lstm"

    # Framework versions (must match SageMaker supported versions)
    pytorch_version: str = "2.5.1"
    python_version: str = "py311"

    @property
    def s3_data_uri(self) -> str:
        return f"s3://{self.bucket}/{self.s3_prefix}/{self.data_prefix}"

    @property
    def s3_model_uri(self) -> str:
        return f"s3://{self.bucket}/{self.s3_prefix}/{self.model_prefix}"

    @property
    def s3_evaluation_uri(self) -> str:
        return f"s3://{self.bucket}/{self.s3_prefix}/{self.evaluation_prefix}"

    @property
    def input_size(self) -> int:
        return len(self.features)

    @property
    def all_tickers(self) -> List[str]:
        return [self.parent_ticker] + self.child_tickers
