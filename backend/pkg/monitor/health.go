package monitor

type HealthStatus struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type ReadinessStatus struct {
	Status   string  `json:"status"`
	Database bool    `json:"database"`
	Redis    *bool   `json:"redis"`
}
