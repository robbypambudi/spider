package apis

type SecurityDecision string

const (
	DecisionAllow  SecurityDecision = "ALLOW"
	DecisionBlock  SecurityDecision = "BLOCK"
	DecisionReview SecurityDecision = "REVIEW"
	DecisionError  SecurityDecision = "ERROR"
)

type SecurityRequest struct {
	RequestID string                 `json:"request_id"`
	Text      string                 `json:"text"`
	Source    *string                `json:"source,omitempty"`
	UserID    *string                `json:"user_id,omitempty"`
	Model     *string                `json:"model,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type DetectionResult struct {
	Detector    string                 `json:"detector"`
	Score       float64                `json:"score"`
	IsInjection bool                   `json:"is_injection"`
	Threshold   *float64               `json:"threshold,omitempty"`
	LatencyMs   float64                `json:"latency_ms"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type AggregatedDetectionResult struct {
	Score            float64           `json:"score"`
	IsInjection      bool              `json:"is_injection"`
	ChunksScanned    int               `json:"chunks_scanned"`
	DetectorResults  []DetectionResult `json:"detector_results"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type SecurityResult struct {
	RequestID        string            `json:"request_id"`
	Decision         SecurityDecision  `json:"decision"`
	Score            float64           `json:"score"`
	DetectorResults  []DetectionResult `json:"detector_results"`
	ChunksScanned    int               `json:"chunks_scanned"`
	TotalLatencyMs   float64           `json:"total_latency_ms"`
	Policy           string            `json:"policy"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type InferenceRequest struct {
	Model           string                 `json:"model"`
	Prompt          string                 `json:"prompt"`
	MaxTokens       int                    `json:"max_tokens,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	SecurityEnabled bool                   `json:"security_enabled"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type InferenceResponse struct {
	RequestID   string                 `json:"request_id"`
	Model       string                 `json:"model"`
	Output      *string                `json:"output,omitempty"`
	FinishReason *string               `json:"finish_reason,omitempty"`
	Usage       map[string]int         `json:"usage,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ProtectedInferenceResponse struct {
	RequestID           string          `json:"request_id"`
	Status              string          `json:"status"`
	Decision            SecurityDecision `json:"decision"`
	Model               string          `json:"model"`
	Output              *string         `json:"output,omitempty"`
	Security            SecurityResult  `json:"security"`
	SecurityOverheadMs  float64         `json:"security_overhead_ms"`
	InferenceLatencyMs  *float64        `json:"inference_latency_ms,omitempty"`
	EndToEndLatencyMs   float64         `json:"end_to_end_latency_ms"`
	WorkerID            *string         `json:"worker_id,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

type GPUResource struct {
	Index          int    `json:"index"`
	Vendor         string `json:"vendor"`
	Name           string `json:"name"`
	MemoryTotalMB  int    `json:"memory_total_mb"`
	MemoryUsedMB   int    `json:"memory_used_mb"`
	Utilization    int    `json:"utilization"`
}

type WorkerResources struct {
	CPUTotal      int           `json:"cpu_total"`
	MemoryTotalMB int           `json:"memory_total_mb"`
	GPUs          []GPUResource `json:"gpus"`
}

type LoadedModel struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type WorkerResource struct {
	WorkerID         string                 `json:"worker_id"`
	Hostname         string                 `json:"hostname"`
	Site             *string                `json:"site,omitempty"`
	Version          string                 `json:"version"`
	Status           string                 `json:"status"`
	Resources        WorkerResources        `json:"resources"`
	Models           []LoadedModel          `json:"models"`
	RunningRequests  int                    `json:"running_requests"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type WorkerHeartbeat struct {
	WorkerID        string                 `json:"worker_id"`
	Status          string                 `json:"status"`
	Resources       *WorkerResources       `json:"resources,omitempty"`
	Models          []LoadedModel          `json:"models"`
	RunningRequests int                    `json:"running_requests"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateWorkerRequest struct {
	Hostname *string                `json:"hostname,omitempty"`
	Site     *string                `json:"site,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Role        string `json:"role"`
	Email       string `json:"email"`
}

type ScanRequest struct {
	Text     string                 `json:"text"`
	Source   *string                `json:"source,omitempty"`
	Model    *string                `json:"model,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type DetectorView struct {
	Detector    string                 `json:"detector"`
	Score       float64                `json:"score"`
	IsInjection bool                   `json:"is_injection"`
	LatencyMs   *float64               `json:"latency_ms,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ScanResponse struct {
	ScanID        *string        `json:"scan_id,omitempty"`
	RequestID     string         `json:"request_id"`
	Decision      string         `json:"decision"`
	Score         float64        `json:"score"`
	ChunksScanned int            `json:"chunks_scanned"`
	Policy        string         `json:"policy"`
	Threshold     *float64       `json:"threshold,omitempty"`
	LatencyMs     float64        `json:"latency_ms"`
	Detectors     []DetectorView `json:"detectors"`
	Model         *string        `json:"model,omitempty"`
}

type InferenceHTTPRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   *int    `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Security    *struct {
		Enabled bool `json:"enabled"`
	} `json:"security,omitempty"`
}

type EvaluateSample struct {
	Text        string `json:"text"`
	IsInjection bool   `json:"is_injection"`
}

type EvaluateRequest struct {
	Samples    []EvaluateSample `json:"samples"`
	Threshold  *float64         `json:"threshold,omitempty"`
	TargetFPR  []float64        `json:"target_fpr,omitempty"`
}

type ConfusionCounts struct {
	TP int `json:"tp"`
	FP int `json:"fp"`
	TN int `json:"tn"`
	FN int `json:"fn"`
}

type EvaluationReport struct {
	Counts          ConfusionCounts    `json:"counts"`
	TPR             float64            `json:"tpr"`
	FPR             float64            `json:"fpr"`
	TNR             float64            `json:"tnr"`
	FNR             float64            `json:"fnr"`
	Precision       float64            `json:"precision"`
	Recall          float64            `json:"recall"`
	F1              float64            `json:"f1"`
	AUC             float64            `json:"auc"`
	TPRAtTargetFPR  map[string]float64 `json:"tpr_at_target_fpr"`
	Threshold       float64            `json:"threshold"`
	Samples         int                `json:"samples"`
}

type PolicyView struct {
	ID                string  `json:"id,omitempty"`
	Name              string  `json:"name"`
	Kind              string  `json:"kind"`
	Threshold         float64 `json:"threshold"`
	ActionOnDetection string  `json:"action_on_detection"`
	Chunker           string  `json:"chunker"`
	ChunkSize         int     `json:"chunk_size"`
	ChunkOverlap      int     `json:"chunk_overlap"`
	IsDefault         bool    `json:"is_default"`
	Status            string  `json:"status,omitempty"`
}

type CreatePolicyRequest struct {
	Name              string  `json:"name"`
	Threshold         float64 `json:"threshold"`
	ActionOnDetection string  `json:"action_on_detection"`
	Chunker           string  `json:"chunker"`
	ChunkSize         int     `json:"chunk_size"`
	ChunkOverlap      int     `json:"chunk_overlap"`
	IsDefault         bool    `json:"is_default"`
}

type UpdatePolicyRequest struct {
	Name              *string  `json:"name,omitempty"`
	Threshold         *float64 `json:"threshold,omitempty"`
	ActionOnDetection *string  `json:"action_on_detection,omitempty"`
	Chunker           *string  `json:"chunker,omitempty"`
	ChunkSize         *int     `json:"chunk_size,omitempty"`
	ChunkOverlap      *int     `json:"chunk_overlap,omitempty"`
	IsDefault         *bool    `json:"is_default,omitempty"`
}

type RuntimeSettingsView struct {
	DefaultDetector string `json:"default_detector"`
	FailMode        string `json:"fail_mode"`
	Chunker         string `json:"chunker"`
	ChunkSize       int    `json:"chunk_size"`
	ChunkOverlap    int    `json:"chunk_overlap"`
}
