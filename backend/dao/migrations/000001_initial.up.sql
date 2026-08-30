CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL,
    is_active BOOLEAN NOT NULL,
    display_name VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE sites (
    id UUID PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE policies (
    id UUID PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    kind VARCHAR(64) NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    action_on_detection VARCHAR(32) NOT NULL,
    config_json TEXT NOT NULL,
    is_default BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE workers (
    id UUID PRIMARY KEY,
    worker_id VARCHAR(128) NOT NULL UNIQUE,
    hostname VARCHAR(255) NOT NULL,
    site VARCHAR(128),
    version VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    cpu_total INTEGER NOT NULL,
    memory_total_mb INTEGER NOT NULL,
    running_requests INTEGER NOT NULL,
    last_heartbeat_at TIMESTAMPTZ,
    models_json TEXT NOT NULL,
    metadata_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_workers_worker_id ON workers (worker_id);

CREATE TABLE worker_gpus (
    id UUID PRIMARY KEY,
    worker_pk UUID NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    worker_id VARCHAR(128) NOT NULL,
    gpu_index INTEGER NOT NULL,
    vendor VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    memory_total_mb INTEGER NOT NULL,
    memory_used_mb INTEGER NOT NULL,
    utilization INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_worker_gpus_worker_id ON worker_gpus (worker_id);

CREATE TABLE worker_heartbeats (
    id UUID PRIMARY KEY,
    worker_pk UUID NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    worker_id VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    payload_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_worker_heartbeats_worker_id ON worker_heartbeats (worker_id);

CREATE TABLE security_scans (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL,
    user_id UUID REFERENCES users(id),
    decision VARCHAR(16) NOT NULL,
    score DOUBLE PRECISION NOT NULL,
    threshold DOUBLE PRECISION,
    detector VARCHAR(64) NOT NULL,
    detector_version VARCHAR(64) NOT NULL,
    policy VARCHAR(64) NOT NULL,
    chunks_scanned INTEGER NOT NULL,
    chunking_strategy VARCHAR(64) NOT NULL,
    latency_ms DOUBLE PRECISION NOT NULL,
    prompt_hash VARCHAR(64) NOT NULL,
    prompt_length INTEGER NOT NULL,
    prompt_text TEXT,
    model_target VARCHAR(255),
    worker_id VARCHAR(128),
    source VARCHAR(64),
    metadata_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_security_scans_request_id ON security_scans (request_id);
CREATE INDEX ix_security_scans_decision ON security_scans (decision);

CREATE TABLE security_chunk_results (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES security_scans(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    detector VARCHAR(64) NOT NULL,
    score DOUBLE PRECISION NOT NULL,
    is_injection BOOLEAN NOT NULL,
    latency_ms DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_security_chunk_results_scan_id ON security_chunk_results (scan_id);

CREATE TABLE detector_executions (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES security_scans(id) ON DELETE CASCADE,
    detector VARCHAR(64) NOT NULL,
    detector_version VARCHAR(64) NOT NULL,
    threshold DOUBLE PRECISION,
    score DOUBLE PRECISION NOT NULL,
    is_injection BOOLEAN NOT NULL,
    latency_ms DOUBLE PRECISION NOT NULL,
    metadata_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_detector_executions_scan_id ON detector_executions (scan_id);

CREATE TABLE inference_requests (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id),
    scan_id UUID REFERENCES security_scans(id),
    model VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL,
    decision VARCHAR(16) NOT NULL,
    worker_id VARCHAR(128),
    end_to_end_latency_ms DOUBLE PRECISION NOT NULL,
    security_overhead_ms DOUBLE PRECISION NOT NULL,
    inference_latency_ms DOUBLE PRECISION,
    output_preview VARCHAR(512),
    metadata_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_inference_requests_request_id ON inference_requests (request_id);
CREATE INDEX ix_inference_requests_status ON inference_requests (status);

CREATE TABLE inference_events (
    id UUID PRIMARY KEY,
    inference_id UUID NOT NULL REFERENCES inference_requests(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    payload_json TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_inference_events_inference_id ON inference_events (inference_id);

CREATE TABLE serving_models (
    id UUID PRIMARY KEY,
    worker_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_serving_models_worker_id ON serving_models (worker_id);

CREATE TABLE serving_nodes (
    id UUID PRIMARY KEY,
    worker_pk UUID NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    worker_id VARCHAR(128) NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
