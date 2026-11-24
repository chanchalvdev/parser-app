package storage

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Azurite's well-known development key; any valid base64 works for signing offline.
const testAccountKey = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="

func TestNewRejectsUnknownBackend(t *testing.T) {
	if _, err := New(Config{Backend: "gcs", Bucket: "b"}); err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestNewAzureRequiresCredentials(t *testing.T) {
	if _, err := New(Config{Backend: BackendAzure, Bucket: "b"}); err == nil {
		t.Fatal("expected error without account/key")
	}
}

func TestAzurePresignPutIsScopedWriteOnlySAS(t *testing.T) {
	store, err := New(Config{
		Backend:          BackendAzure,
		Bucket:           "file-ingestion",
		AzureAccountName: "stparsertest",
		AzureAccountKey:  testAccountKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	up, err := store.PresignPut(context.Background(), "tenants/t1/uploads/u1/a b.zip", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(up.URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "stparsertest.blob.core.windows.net" {
		t.Errorf("host = %s", u.Host)
	}
	if !strings.HasPrefix(u.Path, "/file-ingestion/tenants/t1/uploads/u1/") {
		t.Errorf("path = %s", u.Path)
	}
	q := u.Query()
	if q.Get("sp") != "cw" {
		t.Errorf("permissions sp = %q, want cw (create+write only)", q.Get("sp"))
	}
	if q.Get("spr") != "https" {
		t.Errorf("protocol spr = %q, want https only on real Azure", q.Get("spr"))
	}
	if q.Get("sr") != "b" {
		t.Errorf("resource sr = %q, want b (single blob)", q.Get("sr"))
	}
	if q.Get("sig") == "" || q.Get("se") == "" || q.Get("st") == "" {
		t.Errorf("missing sig/se/st in %s", u.RawQuery)
	}
	if up.Headers["x-ms-blob-type"] != "BlockBlob" {
		t.Errorf("headers = %v", up.Headers)
	}
}

func TestAzurePresignPutRewritesPublicEndpoint(t *testing.T) {
	store, err := New(Config{
		Backend:             BackendAzure,
		Bucket:              "file-ingestion",
		AzureAccountName:    "devstoreaccount1",
		AzureAccountKey:     testAccountKey,
		AzureEndpoint:       "http://azurite:10000/devstoreaccount1",
		AzurePublicEndpoint: "http://localhost:10000",
	})
	if err != nil {
		t.Fatal(err)
	}
	up, err := store.PresignPut(context.Background(), "k.txt", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(up.URL, "http://localhost:10000/devstoreaccount1/file-ingestion/k.txt?") {
		t.Errorf("url = %s", up.URL)
	}
}

func TestRewriteURLHost(t *testing.T) {
	got, err := RewriteURLHost("http://minio:9000/bucket/key?X-Amz-Signature=abc", "localhost:9100")
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost:9100/bucket/key?X-Amz-Signature=abc" {
		t.Errorf("got %s", got)
	}
	got, _ = RewriteURLHost("http://minio:9000/bucket/key", "https://files.example.com/storage/")
	if got != "https://files.example.com/storage/bucket/key" {
		t.Errorf("got %s", got)
	}
	got, _ = RewriteURLHost("http://minio:9000/bucket/key", "")
	if got != "http://minio:9000/bucket/key" {
		t.Errorf("empty endpoint should be a no-op, got %s", got)
	}
}
