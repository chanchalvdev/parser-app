package queue

import (
\t"context"
\t"encoding/json"

\t"github.com/example/file-platform/backend/internal/models"
\t"github.com/redis/go-redis/v9"
)

type Producer struct {
\tClient   *redis.Client
\tQueueName string
}

type Payload models.JobPayload

func NewProducer(addr, password string, db int, queueName string) *Producer {
\trdb := redis.NewClient(&redis.Options{
\t\tAddr:     addr,
\t\tPassword: password,
\t\tDB:       db,
\t})
\tif queueName == "" {
\t\tqueueName = "ingest.jobs"
\t}
\treturn &Producer{Client: rdb, QueueName: queueName}
}

func (p *Producer) Enqueue(ctx context.Context, payload models.JobPayload) error {
\tb, err := json.Marshal(payload)
\tif err != nil {
\t\treturn err
\t}
\treturn p.Client.LPush(ctx, p.QueueName, b).Err()
\t// Worker will block on RPOP.
}
