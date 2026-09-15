package service_test

import (
	"context"
	"errors"
	"testing"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"
	"feature-flag-manager/internal/service"
	"feature-flag-manager/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAndListUseRepository(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	feature := domain.Feature{Name: "flag", Status: domain.StatusOpen}
	repositoryMock.EXPECT().Get(gomock.Any(), "flag").Return(feature, nil)
	repositoryMock.EXPECT().List(gomock.Any()).Return([]domain.Feature{feature}, nil)
	featureService := service.NewFeatureService(repositoryMock, cacheMock)

	actual, err := featureService.Get(context.Background(), "flag")
	require.NoError(t, err)
	assert.Equal(t, feature, actual)
	features, err := featureService.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []domain.Feature{feature}, features)
}

func TestDeleteReportsCacheFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	repositoryMock.EXPECT().Delete(gomock.Any(), "flag").Return(nil)
	cacheMock.EXPECT().Delete(gomock.Any(), "flag").Return(errors.New("redis unavailable"))

	err := service.NewFeatureService(repositoryMock, cacheMock).Delete(context.Background(), "flag")
	assert.ErrorIs(t, err, service.ErrDependency)
}

func TestEvaluateClosedFeatureIsDisabled(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	cacheMock.EXPECT().Get(gomock.Any(), "flag").Return(domain.Feature{Name: "flag", Status: domain.StatusClosed}, nil)

	evaluation, err := service.NewFeatureService(repositoryMock, cacheMock).Evaluate(context.Background(), "flag", "user")
	require.NoError(t, err)
	assert.False(t, evaluation.Enabled)
}

func TestRefreshReportsCacheFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	repositoryMock.EXPECT().List(gomock.Any()).Return(nil, nil)
	cacheMock.EXPECT().Replace(gomock.Any(), nil).Return(errors.New("redis unavailable"))

	_, err := service.NewFeatureService(repositoryMock, cacheMock).RefreshCache(context.Background())
	assert.ErrorIs(t, err, service.ErrDependency)
}

func TestHealthStopsAtPostgresFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	repositoryMock.EXPECT().Ping(gomock.Any()).Return(repository.ErrNotFound)

	err := service.NewFeatureService(repositoryMock, cacheMock).Health(context.Background())
	assert.ErrorIs(t, err, service.ErrDependency)
}
