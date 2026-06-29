package observability

import (
	"go.uber.org/zap"
)

func NewLogger(environment string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	cfg.InitialFields = map[string]any{"env": environment}
	return cfg.Build()
}
