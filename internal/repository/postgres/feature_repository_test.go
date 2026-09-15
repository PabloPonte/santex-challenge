package postgres_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"
	store "feature-flag-manager/internal/repository/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFeatureRepositoryWithPostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.Run(ctx, "postgres:17-alpine",
		testcontainers.WithEnv(map[string]string{"POSTGRES_DB": "feature_flags", "POSTGRES_USER": "feature_flags", "POSTGRES_PASSWORD": "feature_flags"}),
		testcontainers.WithExposedPorts("5432/tcp"), testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, "postgres://feature_flags:feature_flags@"+net.JoinHostPort(host, port.Port())+"/feature_flags?sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	migration, err := os.ReadFile("../../../db/migrations/000001_create_features.up.sql")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)
	featureRepository := store.NewFeatureRepository(pool)
	first := domain.Feature{Name: "first", Description: "one", Status: domain.StatusOpen, StatusDate: time.Now().UTC(), Whitelist: []string{"user"}}
	second := domain.Feature{Name: "second", Status: domain.StatusClosed, StatusDate: first.StatusDate}

	created, err := featureRepository.Create(ctx, first)
	require.NoError(t, err)
	assertFeatureEqual(t, first, created)
	_, err = featureRepository.Create(ctx, first)
	assert.ErrorIs(t, err, repository.ErrConflict)
	_, err = featureRepository.Create(ctx, second)
	require.NoError(t, err)
	features, err := featureRepository.List(ctx)
	require.NoError(t, err)
	require.Len(t, features, 2)
	assertFeatureEqual(t, first, features[0])
	assertFeatureEqual(t, second, features[1])
	first.Description = "updated"
	updated, err := featureRepository.Update(ctx, first)
	require.NoError(t, err)
	assertFeatureEqual(t, first, updated)
	require.NoError(t, featureRepository.Delete(ctx, "first"))
	_, err = featureRepository.Get(ctx, "first")
	assert.ErrorIs(t, err, repository.ErrNotFound)
	err = featureRepository.Delete(ctx, "first")
	assert.ErrorIs(t, err, repository.ErrNotFound)
	require.NoError(t, featureRepository.Ping(ctx))
}

func assertFeatureEqual(t *testing.T, expected, actual domain.Feature) {
	t.Helper()
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.Status, actual.Status)
	assert.True(t, expected.StatusDate.Equal(actual.StatusDate))
	assert.Equal(t, expected.Whitelist, actual.Whitelist)
}
