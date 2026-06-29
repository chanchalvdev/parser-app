package config

import (
\t"os"
\t"testing"
)

func TestLoadDefaults(t *testing.T) {
\t_ = os.Unsetenv("API_LISTEN_ADDR")
\tcfg, err := Load()
\tif err != nil {
\t\tt.Fatalf("unexpected error: %v", err)
\t}
\tif cfg.MaxUploadSizeMB != 250 {
\t\tt.Fatalf("expected default max upload size 250, got %d", cfg.MaxUploadSizeMB)
\t}
\tif cfg.QueueName != "ingest.jobs" {
\t\tt.Fatalf("expected default queue name ingest.jobs, got %s", cfg.QueueName)
\t}
}

