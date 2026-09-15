package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"
)

var (
	ErrValidation = errors.New("validation error")
	ErrDependency = errors.New("dependency error")
	namePattern   = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type FeatureService struct {
	repository repository.FeatureRepository
	cache      repository.FeatureCache
	clock      Clock
}

func NewFeatureService(repository repository.FeatureRepository, cache repository.FeatureCache) *FeatureService {
	return NewFeatureServiceWithClock(repository, cache, systemClock{})
}

func NewFeatureServiceWithClock(repository repository.FeatureRepository, cache repository.FeatureCache, clock Clock) *FeatureService {
	return &FeatureService{repository: repository, cache: cache, clock: clock}
}

type WriteFeature struct {
	Name        string
	Description string
	Status      domain.Status
	Whitelist   []string
}

func (service *FeatureService) Create(ctx context.Context, input WriteFeature) (domain.Feature, error) {
	if err := validateWrite(input, true); err != nil {
		return domain.Feature{}, err
	}
	feature := domain.Feature{
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
		StatusDate:  service.clock.Now(),
		Whitelist:   slices.Clone(input.Whitelist),
	}
	created, err := service.repository.Create(ctx, feature)
	if err != nil {
		return domain.Feature{}, err
	}
	if err := service.cache.Set(ctx, created); err != nil {
		return domain.Feature{}, fmt.Errorf("%w: update redis after create: %v", ErrDependency, err)
	}
	return created, nil
}

func (service *FeatureService) Get(ctx context.Context, name string) (domain.Feature, error) {
	if err := validateName(name); err != nil {
		return domain.Feature{}, err
	}
	return service.repository.Get(ctx, name)
}

func (service *FeatureService) List(ctx context.Context) ([]domain.Feature, error) {
	return service.repository.List(ctx)
}

func (service *FeatureService) Update(ctx context.Context, name string, input WriteFeature) (domain.Feature, error) {
	input.Name = name
	if err := validateWrite(input, true); err != nil {
		return domain.Feature{}, err
	}
	current, err := service.repository.Get(ctx, name)
	if err != nil {
		return domain.Feature{}, err
	}
	statusDate := current.StatusDate
	if current.Status != input.Status {
		statusDate = service.clock.Now()
	}
	feature := domain.Feature{
		Name:        name,
		Description: input.Description,
		Status:      input.Status,
		StatusDate:  statusDate,
		Whitelist:   slices.Clone(input.Whitelist),
	}
	updated, err := service.repository.Update(ctx, feature)
	if err != nil {
		return domain.Feature{}, err
	}
	if err := service.cache.Set(ctx, updated); err != nil {
		return domain.Feature{}, fmt.Errorf("%w: update redis after update: %v", ErrDependency, err)
	}
	return updated, nil
}

func (service *FeatureService) Delete(ctx context.Context, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := service.repository.Delete(ctx, name); err != nil {
		return err
	}
	if err := service.cache.Delete(ctx, name); err != nil {
		return fmt.Errorf("%w: update redis after delete: %v", ErrDependency, err)
	}
	return nil
}

func (service *FeatureService) RefreshCache(ctx context.Context) (int, error) {
	features, err := service.repository.List(ctx)
	if err != nil {
		return 0, err
	}
	if err := service.cache.Replace(ctx, features); err != nil {
		return 0, fmt.Errorf("%w: refresh redis: %v", ErrDependency, err)
	}
	return len(features), nil
}

func (service *FeatureService) Evaluate(ctx context.Context, name, userID string) (domain.Evaluation, error) {
	if err := validateName(name); err != nil {
		return domain.Evaluation{}, err
	}
	if err := validateUserID(userID); err != nil {
		return domain.Evaluation{}, err
	}
	feature, err := service.cache.Get(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.Evaluation{}, err
		}
		return domain.Evaluation{}, fmt.Errorf("%w: read redis: %v", ErrDependency, err)
	}
	enabled := feature.Status == domain.StatusOpen ||
		(feature.Status == domain.StatusWhitelisted && slices.Contains(feature.Whitelist, userID))
	return domain.Evaluation{FeatureName: feature.Name, UserID: userID, Enabled: enabled, Status: feature.Status}, nil
}

func (service *FeatureService) Health(ctx context.Context) error {
	if err := service.repository.Ping(ctx); err != nil {
		return fmt.Errorf("%w: postgres: %v", ErrDependency, err)
	}
	if err := service.cache.Ping(ctx); err != nil {
		return fmt.Errorf("%w: redis: %v", ErrDependency, err)
	}
	return nil
}

func validateWrite(input WriteFeature, includeName bool) error {
	if includeName {
		if err := validateName(input.Name); err != nil {
			return err
		}
	}
	if len(input.Description) > 2000 {
		return fmt.Errorf("%w: description must not exceed 2000 characters", ErrValidation)
	}
	if !input.Status.Valid() {
		return fmt.Errorf("%w: status must be open, closed, or whitelisted", ErrValidation)
	}
	if len(input.Whitelist) > 10000 {
		return fmt.Errorf("%w: whitelist must not exceed 10000 users", ErrValidation)
	}
	users := make(map[string]struct{}, len(input.Whitelist))
	for _, userID := range input.Whitelist {
		if err := validateUserID(userID); err != nil {
			return err
		}
		if _, exists := users[userID]; exists {
			return fmt.Errorf("%w: whitelist contains duplicate userId", ErrValidation)
		}
		users[userID] = struct{}{}
	}
	return nil
}

func validateName(name string) error {
	if !namePattern.MatchString(name) || len(name) > 100 {
		return fmt.Errorf("%w: name must be a lowercase slug of at most 100 characters", ErrValidation)
	}
	return nil
}

func validateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" || len(userID) > 255 {
		return fmt.Errorf("%w: userId must be non-empty and at most 255 characters", ErrValidation)
	}
	return nil
}
