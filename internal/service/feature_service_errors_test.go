package service_test

import (
	"context"
	"errors"
	"testing"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/service"
	"feature-flag-manager/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetRejectsInvalidName(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	_, err := service.NewFeatureService(mocks.NewMockFeatureRepository(controller), mocks.NewMockFeatureCache(controller)).Get(context.Background(), "Invalid")
	assert.ErrorIs(t, err, service.ErrValidation)
}

func TestListAndRefreshPropagateRepositoryFailures(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	repositoryMock.EXPECT().List(gomock.Any()).Return(nil, errors.New("postgres unavailable")).Times(2)
	featureService := service.NewFeatureService(repositoryMock, cacheMock)

	_, listError := featureService.List(context.Background())
	_, refreshError := featureService.RefreshCache(context.Background())
	assert.EqualError(t, listError, "postgres unavailable")
	assert.EqualError(t, refreshError, "postgres unavailable")
}

func TestEvaluateWrapsRedisFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	repositoryMock := mocks.NewMockFeatureRepository(controller)
	cacheMock := mocks.NewMockFeatureCache(controller)
	cacheMock.EXPECT().Get(gomock.Any(), "flag").Return(domain.Feature{}, errors.New("redis unavailable"))

	_, err := service.NewFeatureService(repositoryMock, cacheMock).Evaluate(context.Background(), "flag", "user")
	assert.ErrorIs(t, err, service.ErrDependency)
}
