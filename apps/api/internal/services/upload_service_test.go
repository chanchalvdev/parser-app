package services

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseNumericJSON(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		raw := json.RawMessage(`12.5`)
		got, err := parseNumericJSON(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 12.5 {
			t.Fatalf("expected 12.5, got %v", got)
		}
	})

	t.Run("string", func(t *testing.T) {
		raw := json.RawMessage(`"99"`)
		got, err := parseNumericJSON(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 99 {
			t.Fatalf("expected 99, got %v", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		raw := json.RawMessage(`{bad}`)
		if _, err := parseNumericJSON(raw); err == nil {
			t.Fatal("expected error for invalid JSON numeric value")
		}
	})
}

func TestUploadMetadataParsingAndValidation(t *testing.T) {
	t.Run("valid metadata", func(t *testing.T) {
		raw := json.RawMessage(`{"file_name":" report.txt ","content_type":"text/plain","size_bytes":10}`)
		meta, err := parseUploadMetadata(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(meta.FileName) != "report.txt" {
			t.Fatalf("unexpected filename: %s", meta.FileName)
		}
	})

	t.Run("missing filename", func(t *testing.T) {
		raw := json.RawMessage(`{"size_bytes":10}`)
		if _, err := parseUploadMetadata(raw); err == nil {
			t.Fatal("expected missing file name error")
		}
	})

	t.Run("empty metadata", func(t *testing.T) {
		if _, err := parseUploadMetadata(nil); err == nil {
			t.Fatal("expected error for empty metadata")
		}
	})
}

func TestSanitizeFileName(t *testing.T) {
	t.Run("safe", func(t *testing.T) {
		got, err := sanitizeFileName("hello.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello.txt" {
			t.Fatalf("unexpected name: %s", got)
		}
	})

	t.Run("normalize traversal segments", func(t *testing.T) {
		got, err := sanitizeFileName("../archive/../../evil")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "evil" {
			t.Fatalf("expected '..' segments replaced, got %s", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := sanitizeFileName("   "); err == nil {
			t.Fatal("expected error for blank file name")
		}
		if _, err := sanitizeFileName("/"); err == nil {
			t.Fatal("expected error for root path")
		}
	})
}

func TestValidateUploadMetadata(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if err := validateUploadMetadata(uploadMetadata{FileName: "data.csv"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if err := validateUploadMetadata(uploadMetadata{FileName: "  "}); err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestStoragePrefixAndObjectKey(t *testing.T) {
	prefix := storagePrefix("tenant-a", "upload-1")
	if prefix != "raw/tenant-a/upload-1" {
		t.Fatalf("unexpected prefix: %s", prefix)
	}

	key := storageObjectKey(prefix, "report.txt")
	if key != "raw/tenant-a/upload-1/report.txt" {
		t.Fatalf("unexpected object key: %s", key)
	}
}
