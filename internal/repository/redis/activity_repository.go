package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix   = "activity:"
	maxActions  = 3
	activityTTL = 30 * 24 * time.Hour
)

// ActivityRepository is the only layer that talks to Redis on behalf of the
// per-user activity log.
type ActivityRepository interface {
	RecordAction(ctx context.Context, username, action string) error
	GetActions(ctx context.Context, username string) ([]string, error)
}

type activityRepository struct {
	client *redis.Client
}

// NewActivityRepository builds an ActivityRepository backed by the given
// Redis client.
func NewActivityRepository(client *redis.Client) ActivityRepository {
	return &activityRepository{client: client}
}

func (r *activityRepository) RecordAction(ctx context.Context, username, action string) error {
	key := activityKey(username)

	pipe := r.client.TxPipeline()
	pipe.LPush(ctx, key, action)
	pipe.LTrim(ctx, key, 0, maxActions-1)
	pipe.Expire(ctx, key, activityTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *activityRepository) GetActions(ctx context.Context, username string) ([]string, error) {
	actions, err := r.client.LRange(ctx, activityKey(username), 0, maxActions-1).Result()
	if err != nil {
		return nil, err
	}
	return actions, nil
}

func activityKey(username string) string {
	return fmt.Sprintf("%s%s", keyPrefix, username)
}
