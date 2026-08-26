"""initial spider schema

Revision ID: 0001_initial
Revises:
Create Date: 2026-08-27
"""

from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op

revision: str = "0001_initial"
down_revision: str | None = None
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "users",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("email", sa.String(length=320), nullable=False),
        sa.Column("hashed_password", sa.String(length=255), nullable=False),
        sa.Column("role", sa.String(length=32), nullable=False),
        sa.Column("is_active", sa.Boolean(), nullable=False),
        sa.Column("display_name", sa.String(length=128), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_users"),
        sa.UniqueConstraint("email", name="uq_users_email"),
    )
    op.create_table(
        "sites",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("name", sa.String(length=128), nullable=False),
        sa.Column("description", sa.Text(), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_sites"),
        sa.UniqueConstraint("name", name="uq_sites_name"),
    )
    op.create_table(
        "policies",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("name", sa.String(length=128), nullable=False),
        sa.Column("kind", sa.String(length=64), nullable=False),
        sa.Column("threshold", sa.Float(), nullable=False),
        sa.Column("action_on_detection", sa.String(length=32), nullable=False),
        sa.Column("config_json", sa.Text(), nullable=False),
        sa.Column("is_default", sa.Boolean(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_policies"),
        sa.UniqueConstraint("name", name="uq_policies_name"),
    )
    op.create_table(
        "workers",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("worker_id", sa.String(length=128), nullable=False),
        sa.Column("hostname", sa.String(length=255), nullable=False),
        sa.Column("site", sa.String(length=128), nullable=True),
        sa.Column("version", sa.String(length=32), nullable=False),
        sa.Column("status", sa.String(length=32), nullable=False),
        sa.Column("cpu_total", sa.Integer(), nullable=False),
        sa.Column("memory_total_mb", sa.Integer(), nullable=False),
        sa.Column("running_requests", sa.Integer(), nullable=False),
        sa.Column("last_heartbeat_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("models_json", sa.Text(), nullable=False),
        sa.Column("metadata_json", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_workers"),
        sa.UniqueConstraint("worker_id", name="uq_workers_worker_id"),
    )
    op.create_index("ix_workers_worker_id", "workers", ["worker_id"])
    op.create_table(
        "worker_gpus",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("worker_pk", sa.Uuid(), nullable=False),
        sa.Column("worker_id", sa.String(length=128), nullable=False),
        sa.Column("gpu_index", sa.Integer(), nullable=False),
        sa.Column("vendor", sa.String(length=64), nullable=False),
        sa.Column("name", sa.String(length=255), nullable=False),
        sa.Column("memory_total_mb", sa.Integer(), nullable=False),
        sa.Column("memory_used_mb", sa.Integer(), nullable=False),
        sa.Column("utilization", sa.Integer(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["worker_pk"], ["workers.id"], name="fk_worker_gpus_worker_pk_workers", ondelete="CASCADE"
        ),
        sa.PrimaryKeyConstraint("id", name="pk_worker_gpus"),
    )
    op.create_index("ix_worker_gpus_worker_id", "worker_gpus", ["worker_id"])
    op.create_table(
        "worker_heartbeats",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("worker_pk", sa.Uuid(), nullable=False),
        sa.Column("worker_id", sa.String(length=128), nullable=False),
        sa.Column("status", sa.String(length=32), nullable=False),
        sa.Column("payload_json", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["worker_pk"],
            ["workers.id"],
            name="fk_worker_heartbeats_worker_pk_workers",
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("id", name="pk_worker_heartbeats"),
    )
    op.create_index("ix_worker_heartbeats_worker_id", "worker_heartbeats", ["worker_id"])
    op.create_table(
        "security_scans",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("request_id", sa.Uuid(), nullable=False),
        sa.Column("user_id", sa.Uuid(), nullable=True),
        sa.Column("decision", sa.String(length=16), nullable=False),
        sa.Column("score", sa.Float(), nullable=False),
        sa.Column("threshold", sa.Float(), nullable=True),
        sa.Column("detector", sa.String(length=64), nullable=False),
        sa.Column("detector_version", sa.String(length=64), nullable=False),
        sa.Column("policy", sa.String(length=64), nullable=False),
        sa.Column("chunks_scanned", sa.Integer(), nullable=False),
        sa.Column("chunking_strategy", sa.String(length=64), nullable=False),
        sa.Column("latency_ms", sa.Float(), nullable=False),
        sa.Column("prompt_hash", sa.String(length=64), nullable=False),
        sa.Column("prompt_length", sa.Integer(), nullable=False),
        sa.Column("prompt_text", sa.Text(), nullable=True),
        sa.Column("model_target", sa.String(length=255), nullable=True),
        sa.Column("worker_id", sa.String(length=128), nullable=True),
        sa.Column("source", sa.String(length=64), nullable=True),
        sa.Column("metadata_json", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(["user_id"], ["users.id"], name="fk_security_scans_user_id_users"),
        sa.PrimaryKeyConstraint("id", name="pk_security_scans"),
    )
    op.create_index("ix_security_scans_request_id", "security_scans", ["request_id"])
    op.create_index("ix_security_scans_decision", "security_scans", ["decision"])
    op.create_table(
        "security_chunk_results",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("scan_id", sa.Uuid(), nullable=False),
        sa.Column("chunk_index", sa.Integer(), nullable=False),
        sa.Column("detector", sa.String(length=64), nullable=False),
        sa.Column("score", sa.Float(), nullable=False),
        sa.Column("is_injection", sa.Boolean(), nullable=False),
        sa.Column("latency_ms", sa.Float(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["scan_id"],
            ["security_scans.id"],
            name="fk_security_chunk_results_scan_id_security_scans",
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("id", name="pk_security_chunk_results"),
    )
    op.create_index("ix_security_chunk_results_scan_id", "security_chunk_results", ["scan_id"])
    op.create_table(
        "detector_executions",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("scan_id", sa.Uuid(), nullable=False),
        sa.Column("detector", sa.String(length=64), nullable=False),
        sa.Column("detector_version", sa.String(length=64), nullable=False),
        sa.Column("threshold", sa.Float(), nullable=True),
        sa.Column("score", sa.Float(), nullable=False),
        sa.Column("is_injection", sa.Boolean(), nullable=False),
        sa.Column("latency_ms", sa.Float(), nullable=False),
        sa.Column("metadata_json", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["scan_id"],
            ["security_scans.id"],
            name="fk_detector_executions_scan_id_security_scans",
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("id", name="pk_detector_executions"),
    )
    op.create_index("ix_detector_executions_scan_id", "detector_executions", ["scan_id"])
    op.create_table(
        "inference_requests",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("request_id", sa.Uuid(), nullable=False),
        sa.Column("user_id", sa.Uuid(), nullable=True),
        sa.Column("scan_id", sa.Uuid(), nullable=True),
        sa.Column("model", sa.String(length=255), nullable=False),
        sa.Column("status", sa.String(length=32), nullable=False),
        sa.Column("decision", sa.String(length=16), nullable=False),
        sa.Column("worker_id", sa.String(length=128), nullable=True),
        sa.Column("end_to_end_latency_ms", sa.Float(), nullable=False),
        sa.Column("security_overhead_ms", sa.Float(), nullable=False),
        sa.Column("inference_latency_ms", sa.Float(), nullable=True),
        sa.Column("output_preview", sa.String(length=512), nullable=True),
        sa.Column("metadata_json", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(["user_id"], ["users.id"], name="fk_inference_requests_user_id_users"),
        sa.ForeignKeyConstraint(
            ["scan_id"], ["security_scans.id"], name="fk_inference_requests_scan_id_security_scans"
        ),
        sa.PrimaryKeyConstraint("id", name="pk_inference_requests"),
        sa.UniqueConstraint("request_id", name="uq_inference_requests_request_id"),
    )
    op.create_index("ix_inference_requests_request_id", "inference_requests", ["request_id"])
    op.create_index("ix_inference_requests_status", "inference_requests", ["status"])
    op.create_table(
        "inference_events",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("inference_id", sa.Uuid(), nullable=False),
        sa.Column("event_type", sa.String(length=64), nullable=False),
        sa.Column("payload_json", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["inference_id"],
            ["inference_requests.id"],
            name="fk_inference_events_inference_id_inference_requests",
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("id", name="pk_inference_events"),
    )
    op.create_index("ix_inference_events_inference_id", "inference_events", ["inference_id"])
    op.create_table(
        "serving_models",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("worker_id", sa.String(length=128), nullable=False),
        sa.Column("name", sa.String(length=255), nullable=False),
        sa.Column("status", sa.String(length=32), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.PrimaryKeyConstraint("id", name="pk_serving_models"),
    )
    op.create_index("ix_serving_models_worker_id", "serving_models", ["worker_id"])
    op.create_table(
        "serving_nodes",
        sa.Column("id", sa.Uuid(), nullable=False),
        sa.Column("worker_pk", sa.Uuid(), nullable=False),
        sa.Column("worker_id", sa.String(length=128), nullable=False),
        sa.Column("status", sa.String(length=32), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["worker_pk"], ["workers.id"], name="fk_serving_nodes_worker_pk_workers", ondelete="CASCADE"
        ),
        sa.PrimaryKeyConstraint("id", name="pk_serving_nodes"),
        sa.UniqueConstraint("worker_id", name="uq_serving_nodes_worker_id"),
    )


def downgrade() -> None:
    op.drop_table("serving_nodes")
    op.drop_table("serving_models")
    op.drop_table("inference_events")
    op.drop_table("inference_requests")
    op.drop_table("detector_executions")
    op.drop_table("security_chunk_results")
    op.drop_table("security_scans")
    op.drop_table("worker_heartbeats")
    op.drop_table("worker_gpus")
    op.drop_table("workers")
    op.drop_table("policies")
    op.drop_table("sites")
    op.drop_table("users")
