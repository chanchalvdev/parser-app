package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStore struct {
	client          *minio.Client
	bucket          string
	presignEndpoint string
}

func newMinioStore(cfg Config) (*minioStore, error) {
	endpoint := strings.TrimSpace(cfg.MinioEndpoint)
	client, err := minio.New(
		strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://"),
		&minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
			Secure: strings.HasPrefix(endpoint, "https://"),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &minioStore{client: client, bucket: cfg.Bucket, presignEndpoint: cfg.MinioPresignEndpoint}, nil
}

func (s *minioStore) Backend() string { return BackendMinio }
func (s *minioStore) Bucket() string  { return s.bucket }

func (s *minioStore) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}
	return nil
}

func (s *minioStore) PresignPut(ctx context.Context, objectKey string, expiry time.Duration) (*PresignedUpload, error) {
	presigned, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("create presigned upload url: %w", err)
	}
	// SigV4 signs the Host header, so the presign endpoint must equal what the browser
	// sends; the nginx proxy in front of MinIO preserves it.
	publicURL, err := RewriteURLHost(presigned.String(), s.presignEndpoint)
	if err != nil {
		return nil, fmt.Errorf("rewrite presigned upload url: %w", err)
	}
	return &PresignedUpload{URL: publicURL, Headers: map[string]string{}}, nil
}

func (s *minioStore) Stat(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, fmt.Errorf("%w: %s", ErrObjectNotFound, objectKey)
		}
		return nil, err
	}
	return &ObjectInfo{Size: info.Size}, nil
}
