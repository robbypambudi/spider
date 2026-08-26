export type SecurityDecision = "ALLOW" | "BLOCK" | "REVIEW" | "ERROR";

export interface DetectorView {
  detector: string;
  score: number;
  is_injection: boolean;
  latency_ms?: number;
  metadata?: Record<string, unknown>;
}

export interface ScanResponse {
  scan_id?: string;
  request_id: string;
  decision: SecurityDecision;
  score: number;
  chunks_scanned: number;
  policy: string;
  threshold?: number;
  latency_ms: number;
  detectors: DetectorView[];
  model?: string;
}

export interface ScanListItem {
  id: string;
  request_id: string;
  decision: SecurityDecision;
  score: number;
  threshold?: number;
  detector: string;
  policy: string;
  chunks_scanned: number;
  chunking_strategy: string;
  latency_ms: number;
  prompt_hash: string;
  prompt_length: number;
  model_target?: string;
  created_at: string;
}

export interface ScanDetail extends ScanListItem {
  detector_version: string;
  worker_id?: string;
  chunks: Array<{
    index: number;
    detector: string;
    score: number;
    is_injection: boolean;
    latency_ms: number;
  }>;
  detectors: Array<{
    detector: string;
    version: string;
    score: number;
    is_injection: boolean;
    latency_ms: number;
    threshold?: number;
  }>;
}

export interface MetricsSummary {
  total_scans: number;
  allowed: number;
  blocked: number;
  review: number;
  detection_rate: number;
  avg_detection_latency_ms: number;
  p95_detection_latency_ms: number;
  avg_security_overhead_ms: number;
  active_serving_nodes: number;
  total_gpus: number;
  workers_total: number;
}

export interface WorkerView {
  worker_id: string;
  hostname: string;
  site?: string;
  version: string;
  status: string;
  running_requests: number;
  resources: {
    cpu_total: number;
    memory_total_mb: number;
    gpus: Array<{
      index: number;
      vendor: string;
      name: string;
      memory_total_mb: number;
      memory_used_mb: number;
      utilization: number;
    }>;
  };
  models: Array<{ name: string; status: string }>;
}

export interface InferenceResponse {
  request_id: string;
  status: string;
  decision: SecurityDecision;
  model: string;
  output?: string | null;
  security_overhead_ms: number;
  inference_latency_ms?: number | null;
  end_to_end_latency_ms: number;
  worker_id?: string | null;
  security: {
    decision: SecurityDecision;
    score: number;
    chunks_scanned: number;
    policy: string;
    detector_results: DetectorView[];
  };
}
