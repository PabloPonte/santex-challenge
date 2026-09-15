package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"
	"feature-flag-manager/internal/service"
	"feature-flag-manager/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func TestCreatePersistsAndCachesFeature(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	feature := domain.Feature{Name: "new-checkout", Description: "Roll out checkout", Status: domain.StatusWhitelisted, StatusDate: now, Whitelist: []string{"user-1"}}

	repositoryMock.EXPECT().Create(gomock.Any(), feature).Return(feature, nil)
	cacheMock.EXPECT().Set(gomock.Any(), feature).Return(nil)

	featureService := service.NewFeatureServiceWithClock(repositoryMock, cacheMock, fixedClock{now: now})
	actual, err := featureService.Create(context.Background(), service.WriteFeature{Name: feature.Name, Description: feature.Description, Status: feature.Status, Whitelist: feature.Whitelist})

	require.NoError(t, err)
	assert.Equal(t, feature, actual)
}

func TestUpdateRetainsStatusDateWhenStatusDoesNotChange(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	statusDate := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	current := domain.Feature{Name: "new-checkout", Description: "Old", Status: domain.StatusOpen, StatusDate: statusDate}
	updated := domain.Feature{Name: "new-checkout", Description: "New", Status: domain.StatusOpen, StatusDate: statusDate, Whitelist: []string{"user-1"}}

	repositoryMock.EXPECT().Get(gomock.Any(), "new-checkout").Return(current, nil)
	repositoryMock.EXPECT().Update(gomock.Any(), updated).Return(updated, nil)
	cacheMock.EXPECT().Set(gomock.Any(), updated).Return(nil)

	featureService := service.NewFeatureServiceWithClock(repositoryMock, cacheMock, fixedClock{now: statusDate.Add(time.Hour)})
	actual, err := featureService.Update(context.Background(), "new-checkout", service.WriteFeature{Description: "New", Status: domain.StatusOpen, Whitelist: []string{"user-1"}})

	require.NoError(t, err)
	assert.Equal(t, updated, actual)
}

func TestEvaluateUsesCacheAndWhitelist(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	feature := domain.Feature{Name: "new-checkout", Status: domain.StatusWhitelisted, Whitelist: []string{"allowed-user"}}
	cacheMock.EXPECT().Get(gomock.Any(), "new-checkout").Return(feature, nil)

	featureService := service.NewFeatureService(repositoryMock, cacheMock)
	evaluation, err := featureService.Evaluate(context.Background(), "new-checkout", "allowed-user")

	require.NoError(t, err)
	assert.Equal(t, domain.Evaluation{FeatureName: "new-checkout", UserID: "allowed-user", Enabled: true, Status: domain.StatusWhitelisted}, evaluation)
}

func TestCreateReportsRedisFailureAfterPersistence(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	feature := domain.Feature{Name: "new-checkout", Status: domain.StatusClosed, StatusDate: now}
	repositoryMock.EXPECT().Create(gomock.Any(), feature).Return(feature, nil)
	cacheMock.EXPECT().Set(gomock.Any(), feature).Return(errors.New("redis unavailable"))

	featureService := service.NewFeatureServiceWithClock(repositoryMock, cacheMock, fixedClock{now: now})
	_, err := featureService.Create(context.Background(), service.WriteFeature{Name: feature.Name, Status: feature.Status})

	require.ErrorIs(t, err, service.ErrDependency)
}

func TestEvaluateReturnsNotFoundForCacheMiss(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	cacheMock.EXPECT().Get(gomock.Any(), "new-checkout").Return(domain.Feature{}, repository.ErrNotFound)

	featureService := service.NewFeatureService(repositoryMock, cacheMock)
	_, err := featureService.Evaluate(context.Background(), "new-checkout", "user-1")

	require.ErrorIs(t, err, repository.ErrNotFound)
}
