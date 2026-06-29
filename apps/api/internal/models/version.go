package models

const APIName = "enterprise-file-platform-api"

const APIVersion = "0.1.0"

type HealthStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type VersionResponse struct {
	Service     string `json:"service"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}
