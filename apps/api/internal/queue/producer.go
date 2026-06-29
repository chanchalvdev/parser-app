package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type IngestionJobMessage struct {
	JobID      string `json:"job_id"`
	TenantID   string `json:"tenant_id"`
	RootFileID string `json:"root_file_id"`
	StoragePath string `json:"storage_path"`
	Bucket     string `json:"bucket"`
	OriginalName string `json:"original_name"`
	Depth      int    `json:"depth"`
	CreatedAt  string `json:"created_at"`
}

type Producer interface {
	EnqueueIngestionJob(ctx context.Context, message IngestionJobMessage) error
}

type RedisProducer struct {
	client    *redis.Client
	queueName string
}

func NewRedisProducer(client *redis.Client, queueName string) *RedisProducer {
	if queueName == "" {
		queueName = "ingestion_jobs"
	}

	return &RedisProducer{
		client:    client,
		queueName: queueName,
	}
}

func (p *RedisProducer) EnqueueIngestionJob(ctx context.Context, message IngestionJobMessage) error {
	if p == nil {
		return fmt.Errorf("queue producer is nil")
	}
	if p.client == nil {
		return fmt.Errorf("queue producer has nil redis client")
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal ingestion job message: %w", err)
	}
	return p.client.RPush(ctx, p.queueName, payload).Err()
}
