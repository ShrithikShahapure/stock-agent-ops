"""SageMaker Pipeline definition for stock LSTM model training.

Pipeline steps:
  1. Preprocessing  - Fetch 1yr OHLCV data for ^GSPC + 20 child tickers
  2. Parent Training - Train parent LSTM on ^GSPC (ml.c5.xlarge)
  3. Child Training  - Train all 20 children via transfer learning (ml.c5.xlarge)
  4. Evaluation      - Evaluate all models on validation splits
"""

import sagemaker
from sagemaker.inputs import TrainingInput
from sagemaker.processing import ProcessingInput, ProcessingOutput
from sagemaker.pytorch import PyTorch
from sagemaker.sklearn import SKLearnProcessor
from sagemaker.workflow.parameters import ParameterString, ParameterInteger
from sagemaker.workflow.pipeline import Pipeline
from sagemaker.workflow.steps import ProcessingStep, TrainingStep

from config import SageMakerConfig


def create_pipeline(cfg: SageMakerConfig, sagemaker_session: sagemaker.Session) -> Pipeline:
    """Build and return the SageMaker Pipeline."""

    role = cfg.role_arn
    bucket = cfg.bucket

    # ── Pipeline parameters (overridable at execution time) ──────────────

    p_instance_type = ParameterString(
        name="TrainingInstanceType", default_value=cfg.training_instance_type)
    p_epochs_parent = ParameterInteger(
        name="ParentEpochs", default_value=cfg.parent_epochs)
    p_epochs_child = ParameterInteger(
        name="ChildEpochs", default_value=cfg.child_epochs)
    p_batch_size = ParameterInteger(
        name="BatchSize", default_value=cfg.batch_size)
    p_lookback = ParameterInteger(
        name="LookbackDays", default_value=cfg.lookback_days)

    all_tickers = ",".join(cfg.all_tickers)
    child_tickers = ",".join(cfg.child_tickers)

    # S3 paths
    s3_data = f"s3://{bucket}/{cfg.s3_prefix}/{cfg.data_prefix}"
    s3_parent_model = f"s3://{bucket}/{cfg.s3_prefix}/{cfg.model_prefix}/parent"
    s3_child_model = f"s3://{bucket}/{cfg.s3_prefix}/{cfg.model_prefix}/children"
    s3_evaluation = f"s3://{bucket}/{cfg.s3_prefix}/{cfg.evaluation_prefix}"

    # ── Step 1: Preprocessing ────────────────────────────────────────────

    sklearn_processor = SKLearnProcessor(
        framework_version="1.2-1",
        role=role,
        instance_type=cfg.processing_instance_type,
        instance_count=cfg.processing_instance_count,
        max_runtime_in_seconds=cfg.processing_timeout,
        sagemaker_session=sagemaker_session,
    )

    step_preprocess = ProcessingStep(
        name="FetchOHLCVData",
        processor=sklearn_processor,
        code="scripts/preprocess.py",
        outputs=[
            ProcessingOutput(
                output_name="data",
                source="/opt/ml/processing/output",
                destination=s3_data,
            ),
        ],
        job_arguments=[
            "--tickers", all_tickers,
            "--lookback-days", str(cfg.lookback_days),
        ],
    )

    # ── Step 2: Parent training ──────────────────────────────────────────

    parent_estimator = PyTorch(
        entry_point="train_parent.py",
        source_dir="scripts",
        role=role,
        instance_count=cfg.training_instance_count,
        instance_type=p_instance_type,
        framework_version=cfg.pytorch_version,
        py_version=cfg.python_version,
        volume_size=cfg.training_volume_size_gb,
        max_run=cfg.training_timeout,
        output_path=s3_parent_model,
        sagemaker_session=sagemaker_session,
        hyperparameters={
            "epochs": p_epochs_parent,
            "batch-size": p_batch_size,
            "lr": 1e-3,
            "context-len": cfg.context_len,
            "pred-len": cfg.pred_len,
            "hidden-size": cfg.hidden_size,
            "parent-ticker": cfg.parent_ticker,
        },
    )

    step_train_parent = TrainingStep(
        name="TrainParentModel",
        estimator=parent_estimator,
        inputs={
            "training": TrainingInput(
                s3_data=step_preprocess.properties.ProcessingOutputConfig.Outputs[
                    "data"
                ].S3Output.S3Uri,
                content_type="text/csv",
            ),
        },
    )

    # ── Step 3: Child training ───────────────────────────────────────────

    child_estimator = PyTorch(
        entry_point="train_children.py",
        source_dir="scripts",
        role=role,
        instance_count=cfg.training_instance_count,
        instance_type=p_instance_type,
        framework_version=cfg.pytorch_version,
        py_version=cfg.python_version,
        volume_size=cfg.training_volume_size_gb,
        max_run=cfg.training_timeout,
        output_path=s3_child_model,
        sagemaker_session=sagemaker_session,
        hyperparameters={
            "child-tickers": child_tickers,
            "parent-ticker": cfg.parent_ticker,
            "epochs": p_epochs_child,
            "batch-size": p_batch_size,
            "lr": 3e-4,
            "context-len": cfg.context_len,
            "pred-len": cfg.pred_len,
            "hidden-size": cfg.hidden_size,
            "transfer-strategy": cfg.transfer_strategy,
        },
    )

    step_train_children = TrainingStep(
        name="TrainChildModels",
        estimator=child_estimator,
        inputs={
            "training": TrainingInput(
                s3_data=step_preprocess.properties.ProcessingOutputConfig.Outputs[
                    "data"
                ].S3Output.S3Uri,
                content_type="text/csv",
            ),
            "parent": TrainingInput(
                s3_data=step_train_parent.properties.ModelArtifacts.S3ModelArtifacts,
                content_type="application/x-tar",
            ),
        },
    )

    # ── Step 4: Evaluation ───────────────────────────────────────────────

    eval_processor = SKLearnProcessor(
        framework_version="1.2-1",
        role=role,
        instance_type=cfg.processing_instance_type,
        instance_count=1,
        max_runtime_in_seconds=cfg.processing_timeout,
        sagemaker_session=sagemaker_session,
    )

    step_evaluate = ProcessingStep(
        name="EvaluateModels",
        processor=eval_processor,
        code="scripts/evaluate.py",
        inputs=[
            ProcessingInput(
                source=step_preprocess.properties.ProcessingOutputConfig.Outputs[
                    "data"
                ].S3Output.S3Uri,
                destination="/opt/ml/processing/input/data",
                input_name="data",
            ),
            ProcessingInput(
                source=step_train_parent.properties.ModelArtifacts.S3ModelArtifacts,
                destination="/opt/ml/processing/input/parent_model",
                input_name="parent_model",
            ),
            ProcessingInput(
                source=step_train_children.properties.ModelArtifacts.S3ModelArtifacts,
                destination="/opt/ml/processing/input/child_model",
                input_name="child_model",
            ),
        ],
        outputs=[
            ProcessingOutput(
                output_name="evaluation",
                source="/opt/ml/processing/output",
                destination=s3_evaluation,
            ),
        ],
        job_arguments=[
            "--parent-ticker", cfg.parent_ticker,
            "--child-tickers", child_tickers,
            "--context-len", str(cfg.context_len),
            "--pred-len", str(cfg.pred_len),
            "--hidden-size", str(cfg.hidden_size),
        ],
    )

    # ── Pipeline ─────────────────────────────────────────────────────────

    pipeline = Pipeline(
        name=cfg.pipeline_name,
        parameters=[
            p_instance_type,
            p_epochs_parent,
            p_epochs_child,
            p_batch_size,
            p_lookback,
        ],
        steps=[
            step_preprocess,
            step_train_parent,
            step_train_children,
            step_evaluate,
        ],
        sagemaker_session=sagemaker_session,
    )

    return pipeline
