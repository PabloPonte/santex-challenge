package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeatureRepository struct {
	pool *pgxpool.Pool
}

func NewFeatureRepository(pool *pgxpool.Pool) *FeatureRepository {
	return &FeatureRepository{pool: pool}
}

func (repository *FeatureRepository) Create(ctx context.Context, feature domain.Feature) (domain.Feature, error) {
	whitelist, err := json.Marshal(feature.Whitelist)
	if err != nil {
		return domain.Feature{}, fmt.Errorf("marshal whitelist: %w", err)
	}
	row := repository.pool.QueryRow(ctx, `
		INSERT INTO features (name, description, status, status_date, whitelist)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING name, description, status, status_date, whitelist`,
		feature.Name, feature.Description, feature.Status, feature.StatusDate, whitelist)
	return scanFeature(row)
}

func (repository *FeatureRepository) Get(ctx context.Context, name string) (domain.Feature, error) {
	row := repository.pool.QueryRow(ctx, `
		SELECT name, description, status, status_date, whitelist FROM features WHERE name = $1`, name)
	return scanFeature(row)
}

func (repository *FeatureRepository) List(ctx context.Context) ([]domain.Feature, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT name, description, status, status_date, whitelist FROM features ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}
	defer rows.Close()

	features := make([]domain.Feature, 0)
	for rows.Next() {
		feature, err := scanFeature(rows)
		if err != nil {
			return nil, err
		}
		features = append(features, feature)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate features: %w", err)
	}
	return features, nil
}

func (repository *FeatureRepository) Update(ctx context.Context, feature domain.Feature) (domain.Feature, error) {
	whitelist, err := json.Marshal(feature.Whitelist)
	if err != nil {
		return domain.Feature{}, fmt.Errorf("marshal whitelist: %w", err)
	}
	row := repository.pool.QueryRow(ctx, `
		UPDATE features
		SET description = $2, status = $3, status_date = $4, whitelist = $5
		WHERE name = $1
		RETURNING name, description, status, status_date, whitelist`,
		feature.Name, feature.Description, feature.Status, feature.StatusDate, whitelist)
	return scanFeature(row)
}

func (store *FeatureRepository) Delete(ctx context.Context, name string) error {
	command, err := store.pool.Exec(ctx, `DELETE FROM features WHERE name = $1`, name)
	if err != nil {
		return fmt.Errorf("delete feature: %w", err)
	}
	if command.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (repository *FeatureRepository) Ping(ctx context.Context) error {
	return repository.pool.Ping(ctx)
}

type rowScanner interface {
	Scan(...any) error
}

func scanFeature(row rowScanner) (domain.Feature, error) {
	var feature domain.Feature
	var whitelist []byte
	if err := row.Scan(&feature.Name, &feature.Description, &feature.Status, &feature.StatusDate, &whitelist); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Feature{}, repository.ErrNotFound
		}
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return domain.Feature{}, repository.ErrConflict
		}
		return domain.Feature{}, fmt.Errorf("scan feature: %w", err)
	}
	if err := json.Unmarshal(whitelist, &feature.Whitelist); err != nil {
		return domain.Feature{}, fmt.Errorf("unmarshal whitelist: %w", err)
	}
	return feature, nil
}
