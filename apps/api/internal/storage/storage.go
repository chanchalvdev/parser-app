// Package storage hides which object store holds uploaded files.
//
// Two backends exist: MinIO/S3 (local stack, Render + Cloudflare R2) and Azure Blob Storage.
// The API only needs three operations from either: make sure the bucket/container exists,
// hand the browser a short-lived upload URL, and confirm the object arrived.
package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	BackendMinio = "minio"
	BackendAzure = "azure"
)

// ErrObjectNotFound is returned by Stat when the object does not exist.
var ErrObjectNotFound = errors.New("object not found")

// PresignedUpload is everything a client needs to PUT a file directly into storage.
type PresignedUpload struct {
	URL string
	// Headers the client must send with the PUT. Azure requires x-ms-blob-type;
	// S3/MinIO requires none.
	Headers map[string]string
}

// ObjectInfo is the subset of object metadata the API uses.
type ObjectInfo struct {
	Size int64
}

type ObjectStore interface {
	// Backend returns BackendMinio or BackendAzure.
	Backend() string
	// Bucket is the bucket (MinIO) or container (Azure) name; it is sent to the worker in queue messages.
	Bucket() string
	EnsureBucket(ctx context.Context) error
	PresignPut(ctx context.Context, objectKey string, expiry time.Duration) (*PresignedUpload, error)
	Stat(ctx context.Context, objectKey string) (*ObjectInfo, error)
}

// Config selects and configures a backend. Only the fields of the chosen backend are used.
type Config struct {
	Backend string
	Bucket  string

	MinioEndpoint        string
	MinioPresignEndpoint string
	MinioAccessKey       string
	MinioSecretKey       string

	AzureAccountName string
	AzureAccountKey  string
	// AzureEndpoint overrides https://<account>.blob.core.windows.net (e.g. Azurite).
	AzureEndpoint string
	// AzurePublicEndpoint rewrites the host of SAS URLs handed to browsers, for when the
	// API reaches storage at a different address than the browser does (Azurite in compose).
	AzurePublicEndpoint string
}

func New(cfg Config) (ObjectStore, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, errors.New("storage bucket/container name is required")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Backend)) {
	case "", BackendMinio:
		return newMinioStore(cfg)
	case BackendAzure:
		return newAzureStore(cfg)
	default:
		return nil, fmt.Errorf("unknown STORAGE_BACKEND %q (want %q or %q)", cfg.Backend, BackendMinio, BackendAzure)
	}
}

// RewriteURLHost swaps the scheme and host of rawURL for those of publicEndpoint, keeping the
// path and query (and therefore the signature) intact. A path on publicEndpoint is prefixed.
func RewriteURLHost(rawURL, publicEndpoint string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	publicEndpoint = strings.TrimSpace(publicEndpoint)
	if rawURL == "" {
		return "", errors.New("raw presigned url is required")
	}
	if publicEndpoint == "" {
		return rawURL, nil
	}
	if !strings.Contains(publicEndpoint, "://") {
		publicEndpoint = "http://" + publicEndpoint
	}
	public, err := url.Parse(publicEndpoint)
	if err != nil {
		return "", err
	}
	original, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	original.Scheme = public.Scheme
	original.Host = public.Host
	if public.Path != "" && public.Path != "/" {
		original.Path = strings.TrimSuffix(public.Path, "/") + "/" + strings.TrimPrefix(original.Path, "/")
	}
	return original.String(), nil
}
