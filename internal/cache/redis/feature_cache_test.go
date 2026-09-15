package redis_test

import (
	"context"
	"net"
	"testing"
	"time"

	cache "feature-flag-manager/internal/cache/redis"
	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFeatureCacheWithRedis(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.Run(ctx, "redis:8-alpine", testcontainers.WithExposedPorts("6379/tcp"), testcontainers.WithWaitStrategy(wait.ForListeningPort("6379/tcp")))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, port.Port())})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	featureCache := cache.NewFeatureCache(client)
	feature := domain.Feature{Name: "first", Status: domain.StatusWhitelisted, StatusDate: time.Now().UTC(), Whitelist: []string{"user"}}

	require.NoError(t, featureCache.Set(ctx, feature))
	actual, err := featureCache.Get(ctx, "first")
	require.NoError(t, err)
	assert.Equal(t, feature, actual)
	require.NoError(t, featureCache.Replace(ctx, []domain.Feature{{Name: "second", Status: domain.StatusOpen}}))
	_, err = featureCache.Get(ctx, "first")
	assert.ErrorIs(t, err, repository.ErrNotFound)
	_, err = featureCache.Get(ctx, "second")
	require.NoError(t, err)
	require.NoError(t, featureCache.Delete(ctx, "second"))
	_, err = featureCache.Get(ctx, "second")
	assert.ErrorIs(t, err, repository.ErrNotFound)
	require.NoError(t, featureCache.Ping(ctx))
	require.NoError(t, client.Set(ctx, "feature:broken", "not-json", 0).Err())
	_, err = featureCache.Get(ctx, "broken")
	assert.Error(t, err)
}
