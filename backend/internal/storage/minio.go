package storage

import (
\t"context"
\t"io"

\t"github.com/minio/minio-go/v7"
\t"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config interface {
\tMinioEndpoint() string
\tMinioAccessKey() string
\tMinioSecretKey() string
\tMinioBucket() string
\tMinioUseSSL() bool
}

type Client struct {
\tMinio *minio.Client
}

func New(endpoint, accessKey, secretKey string, useSSL bool) (*Client, error) {
\tcli, err := minio.New(endpoint, &minio.Options{
\t\tCreds: credentials.NewStaticV4(accessKey, secretKey, ""),
\t\tSecure: useSSL,
\t})
\tif err != nil {
\t\treturn nil, err
\t}
\treturn &Client{Minio: cli}, nil
}

func (c *Client) EnsureBucket(ctx context.Context, bucket string) error {
\texists, err := c.Minio.BucketExists(ctx, bucket)
\tif err != nil {
\t\treturn err
\t}
\tif !exists {
\t\treturn c.Minio.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
\t}
\treturn nil
}

func (c *Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
\t_, err := c.Minio.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
\t\tContentType: contentType,
\t})
\treturn err
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) (io.Reader, error) {
\tobj, err := c.Minio.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
\tif err != nil {
\t\treturn nil, err
\t}
\treturn obj, nil
}

