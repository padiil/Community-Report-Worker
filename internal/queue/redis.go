package queue

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func BLPop(client *redis.Client, ctx context.Context, queueName string) ([]string, error) {
	return client.BLPop(ctx, 0, queueName).Result()
}
