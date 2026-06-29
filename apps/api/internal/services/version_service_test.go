package services

import "testing"

func TestVersionServiceGetVersion(t *testing.T) {
	service := NewVersionService()
	result := service.GetVersion("unit-test")

	if result.Service != "enterprise-file-platform-api" {
		t.Fatalf("unexpected service name: %s", result.Service)
	}
	if result.Version != "v0.1.0" {
		t.Fatalf("unexpected version: %s", result.Version)
	}
	if result.Environment != "unit-test" {
		t.Fatalf("unexpected environment: %s", result.Environment)
	}
}

