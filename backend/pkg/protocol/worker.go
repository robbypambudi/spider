package protocol

const (
	WorkerTokenHeader = "X-Spider-Worker-Token"
	RegisterPath      = "/api/v1/workers/register"
	HeartbeatPath     = "/api/v1/workers/{worker_id}/heartbeat"
)
