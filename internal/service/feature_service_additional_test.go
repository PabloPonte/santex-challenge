package service_test

import (
	"context"
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

func TestUpdateChangesStatusDateWhenStatusChanges(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	oldDate := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	newDate := oldDate.Add(time.Hour)
	current := domain.Feature{Name: "flag", Status: domain.StatusClosed, StatusDate: oldDate}
	updated := domain.Feature{Name: "flag", Status: domain.StatusOpen, StatusDate: newDate}
	repositoryMock.EXPECT().Get(gomock.Any(), "flag").Return(current, nil)
	repositoryMock.EXPECT().Update(gomock.Any(), updated).Return(updated, nil)
	cacheMock.EXPECT().Set(gomock.Any(), updated).Return(nil)

	actual, err := service.NewFeatureServiceWithClock(repositoryMock, cacheMock, fixedClock{now: newDate}).Update(context.Background(), "flag", service.WriteFeature{Status: domain.StatusOpen})
	require.NoError(t, err)
	assert.Equal(t, newDate, actual.StatusDate)
}

func TestRefreshCacheReplacesAllFeatures(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	features := []domain.Feature{{Name: "one", Status: domain.StatusOpen}, {Name: "two", Status: domain.StatusClosed}}
	repositoryMock.EXPECT().List(gomock.Any()).Return(features, nil)
	cacheMock.EXPECT().Replace(gomock.Any(), features).Return(nil)

	refreshed, err := service.NewFeatureService(repositoryMock, cacheMock).RefreshCache(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, refreshed)
}

func TestDeleteRemovesCachedFeature(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	repositoryMock.EXPECT().Delete(gomock.Any(), "flag").Return(nil)
	cacheMock.EXPECT().Delete(gomock.Any(), "flag").Return(nil)

	err := service.NewFeatureService(repositoryMock, cacheMock).Delete(context.Background(), "flag")
	require.NoError(t, err)
}

func TestCreateRejectsInvalidInputBeforeDependencies(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	featureService := service.NewFeatureService(mocks.NewMockFeatureRepository(controller), mocks.NewMockFeatureCache(controller))
	_, err := featureService.Create(context.Background(), service.WriteFeature{Name: "Invalid name", Status: domain.StatusOpen})
	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestHealthReportsDependencyFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	repositoryMock.EXPECT().Ping(gomock.Any()).Return(nil)
	cacheMock.EXPECT().Ping(gomock.Any()).Return(repository.ErrNotFound)

	err := service.NewFeatureService(repositoryMock, cacheMock).Health(context.Background())
	assert.ErrorIs(t, err, service.ErrDependency)
}
