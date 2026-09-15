package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "feature:"

type FeatureCache struct {
	client *redis.Client
}

func NewFeatureCache(client *redis.Client) *FeatureCache {
	return &FeatureCache{client: client}
}

func (cache *FeatureCache) Set(ctx context.Context, feature domain.Feature) error {
	payload, err := json.Marshal(feature)
	if err != nil {
		return fmt.Errorf("marshal feature: %w", err)
	}
	if err := cache.client.Set(ctx, key(feature.Name), payload, 0).Err(); err != nil {
		return fmt.Errorf("cache feature: %w", err)
	}
	return nil
}

func (cache *FeatureCache) Get(ctx context.Context, name string) (domain.Feature, error) {
	payload, err := cache.client.Get(ctx, key(name)).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.Feature{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.Feature{}, fmt.Errorf("get cached feature: %w", err)
	}
	var feature domain.Feature
	if err := json.Unmarshal(payload, &feature); err != nil {
		return domain.Feature{}, fmt.Errorf("decode cached feature: %w", err)
	}
	return feature, nil
}

func (cache *FeatureCache) Delete(ctx context.Context, name string) error {
	if err := cache.client.Del(ctx, key(name)).Err(); err != nil {
		return fmt.Errorf("delete cached feature: %w", err)
	}
	return nil
}

func (cache *FeatureCache) Replace(ctx context.Context, features []domain.Feature) error {
	keys, err := cache.client.Keys(ctx, keyPrefix+"*").Result()
	if err != nil {
		return fmt.Errorf("list cache keys: %w", err)
	}
	pipeline := cache.client.TxPipeline()
	if len(keys) > 0 {
		pipeline.Del(ctx, keys...)
	}
	for _, feature := range features {
		payload, err := json.Marshal(feature)
		if err != nil {
			return fmt.Errorf("marshal feature: %w", err)
		}
		pipeline.Set(ctx, key(feature.Name), payload, 0)
	}
	if _, err := pipeline.Exec(ctx); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}
	return nil
}

func (cache *FeatureCache) Ping(ctx context.Context) error {
	return cache.client.Ping(ctx).Err()
}

func key(name string) string {
	return keyPrefix + name
}
