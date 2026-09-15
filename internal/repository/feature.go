package repository

import (
	"context"
	"errors"

	"feature-flag-manager/internal/domain"
)

var (
	ErrNotFound = errors.New("feature not found")
	ErrConflict = errors.New("feature already exists")
)

type FeatureRepository interface {
	Create(context.Context, domain.Feature) (domain.Feature, error)
	Get(context.Context, string) (domain.Feature, error)
	List(context.Context) ([]domain.Feature, error)
	Update(context.Context, domain.Feature) (domain.Feature, error)
	Delete(context.Context, string) error
	Ping(context.Context) error
}

type FeatureCache interface {
	Set(context.Context, domain.Feature) error
	Get(context.Context, string) (domain.Feature, error)
	Delete(context.Context, string) error
	Replace(context.Context, []domain.Feature) error
	Ping(context.Context) error
}
