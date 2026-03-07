#!/usr/bin/env python3
"""CLI to create, start, and inspect the SageMaker training pipeline.

Usage:
    python sagemaker/run_pipeline.py create     # Create/update the pipeline
    python sagemaker/run_pipeline.py start      # Start a pipeline execution
    python sagemaker/run_pipeline.py status     # Check latest execution status
    python sagemaker/run_pipeline.py list       # List recent executions
    python sagemaker/run_pipeline.py describe   # Describe pipeline definition
"""

import argparse
import json
import sys
import time

import boto3
import sagemaker

from config import SageMakerConfig
from pipeline import create_pipeline


def get_session_and_config():
    cfg = SageMakerConfig()

    boto_session = boto3.Session(region_name=cfg.region)
    sm_session = sagemaker.Session(boto_session=boto_session)

    if not cfg.bucket:
        cfg.bucket = sm_session.default_bucket()
        print(f"Using default SageMaker bucket: {cfg.bucket}")

    if not cfg.role_arn:
        cfg.role_arn = sagemaker.get_execution_role()
        print(f"Using execution role: {cfg.role_arn}")

    return cfg, sm_session


def cmd_create(args):
    """Create or update the pipeline."""
    cfg, session = get_session_and_config()
    pipeline = create_pipeline(cfg, session)

    response = pipeline.upsert(role_arn=cfg.role_arn)
    print(f"Pipeline created/updated: {response['PipelineArn']}")
    return response


def cmd_start(args):
    """Start a pipeline execution."""
    cfg, session = get_session_and_config()
    pipeline = create_pipeline(cfg, session)

    params = {}
    if args.instance_type:
        params["TrainingInstanceType"] = args.instance_type
    if args.parent_epochs:
        params["ParentEpochs"] = args.parent_epochs
    if args.child_epochs:
        params["ChildEpochs"] = args.child_epochs
    if args.lookback_days:
        params["LookbackDays"] = args.lookback_days

    execution = pipeline.start(
        execution_display_name=args.name or None,
        parameters=params if params else None,
    )
    print(f"Pipeline execution started: {execution.arn}")

    if args.wait:
        print("Waiting for pipeline to complete...")
        execution.wait(delay=30, max_attempts=240)  # up to 2 hours
        status = execution.describe()["PipelineExecutionStatus"]
        print(f"Pipeline execution finished: {status}")
        if status != "Succeeded":
            sys.exit(1)

    return execution


def cmd_status(args):
    """Check status of the latest or specified execution."""
    cfg, _ = get_session_and_config()
    sm_client = boto3.client("sagemaker", region_name=cfg.region)

    if args.execution_arn:
        response = sm_client.describe_pipeline_execution(
            PipelineExecutionArn=args.execution_arn)
    else:
        executions = sm_client.list_pipeline_executions(
            PipelineName=cfg.pipeline_name,
            MaxResults=1,
            SortBy="CreationTime",
            SortOrder="Descending",
        )
        if not executions["PipelineExecutionSummaries"]:
            print("No executions found")
            return
        exec_arn = executions["PipelineExecutionSummaries"][0]["PipelineExecutionArn"]
        response = sm_client.describe_pipeline_execution(PipelineExecutionArn=exec_arn)

    print(json.dumps({
        "PipelineExecutionArn": response["PipelineExecutionArn"],
        "Status": response["PipelineExecutionStatus"],
        "CreationTime": str(response.get("CreationTime", "")),
        "LastModifiedTime": str(response.get("LastModifiedTime", "")),
    }, indent=2))

    # Show step statuses
    steps = sm_client.list_pipeline_execution_steps(
        PipelineExecutionArn=response["PipelineExecutionArn"])
    print("\nSteps:")
    for step in steps["PipelineExecutionSteps"]:
        status = step["StepStatus"]
        name = step["StepName"]
        print(f"  {name}: {status}")


def cmd_list(args):
    """List recent pipeline executions."""
    cfg, _ = get_session_and_config()
    sm_client = boto3.client("sagemaker", region_name=cfg.region)

    executions = sm_client.list_pipeline_executions(
        PipelineName=cfg.pipeline_name,
        MaxResults=args.max_results,
        SortBy="CreationTime",
        SortOrder="Descending",
    )

    for ex in executions["PipelineExecutionSummaries"]:
        print(f"  {ex['PipelineExecutionStatus']:12s}  "
              f"{str(ex.get('StartTime', '')):25s}  "
              f"{ex['PipelineExecutionArn']}")


def cmd_describe(args):
    """Describe the pipeline definition."""
    cfg, _ = get_session_and_config()
    sm_client = boto3.client("sagemaker", region_name=cfg.region)

    try:
        response = sm_client.describe_pipeline(PipelineName=cfg.pipeline_name)
        print(json.dumps({
            "PipelineName": response["PipelineName"],
            "PipelineArn": response["PipelineArn"],
            "Status": response["PipelineStatus"],
            "CreationTime": str(response.get("CreationTime", "")),
            "LastModifiedTime": str(response.get("LastModifiedTime", "")),
            "RoleArn": response.get("RoleArn", ""),
        }, indent=2))
    except sm_client.exceptions.ResourceNotFound:
        print(f"Pipeline '{cfg.pipeline_name}' not found. Run 'create' first.")
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="SageMaker Training Pipeline CLI")
    subparsers = parser.add_subparsers(dest="command", required=True)

    # create
    subparsers.add_parser("create", help="Create or update the pipeline")

    # start
    sp_start = subparsers.add_parser("start", help="Start a pipeline execution")
    sp_start.add_argument("--name", type=str, help="Execution display name")
    sp_start.add_argument("--wait", action="store_true", help="Wait for completion")
    sp_start.add_argument("--instance-type", type=str, help="Override training instance type")
    sp_start.add_argument("--parent-epochs", type=int)
    sp_start.add_argument("--child-epochs", type=int)
    sp_start.add_argument("--lookback-days", type=int)

    # status
    sp_status = subparsers.add_parser("status", help="Check execution status")
    sp_status.add_argument("--execution-arn", type=str)

    # list
    sp_list = subparsers.add_parser("list", help="List recent executions")
    sp_list.add_argument("--max-results", type=int, default=10)

    # describe
    subparsers.add_parser("describe", help="Describe pipeline definition")

    args = parser.parse_args()

    commands = {
        "create": cmd_create,
        "start": cmd_start,
        "status": cmd_status,
        "list": cmd_list,
        "describe": cmd_describe,
    }
    commands[args.command](args)


if __name__ == "__main__":
    main()
