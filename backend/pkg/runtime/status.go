package runtime

const (
	WorkerStatusRegistering = "REGISTERING"
	WorkerStatusOnline      = "ONLINE"
	WorkerStatusBusy        = "BUSY"
	WorkerStatusDraining    = "DRAINING"
	WorkerStatusOffline     = "OFFLINE"
	WorkerStatusError       = "ERROR"

	InferenceStatusPending   = "pending"
	InferenceStatusScanning  = "scanning"
	InferenceStatusBlocked   = "blocked"
	InferenceStatusReview    = "review"
	InferenceStatusRouting   = "routing"
	InferenceStatusCompleted = "completed"
	InferenceStatusFailed    = "failed"

	ModelStatusLoading   = "LOADING"
	ModelStatusReady     = "READY"
	ModelStatusUnloading = "UNLOADING"
	ModelStatusError     = "ERROR"
)
