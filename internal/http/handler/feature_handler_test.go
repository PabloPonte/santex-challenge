package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/http/handler"
	"feature-flag-manager/internal/http/handler/mocks"
	"feature-flag-manager/internal/http/router"
	"feature-flag-manager/internal/repository"
	"feature-flag-manager/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFeatureRoutes(t *testing.T) {
	controller := gomock.NewController(t)
	useCase := mocks.NewMockFeatureUseCase(controller)
	api := router.New(handler.NewFeatureHandler(useCase))
	feature := domain.Feature{Name: "flag", Description: "test", Status: domain.StatusOpen}
	useCase.EXPECT().Create(gomock.Any(), service.WriteFeature{Name: "flag", Description: "test", Status: domain.StatusOpen, Whitelist: []string{}}).Return(feature, nil)
	useCase.EXPECT().Get(gomock.Any(), "missing").Return(domain.Feature{}, repository.ErrNotFound)
	useCase.EXPECT().Delete(gomock.Any(), "flag").Return(nil)
	useCase.EXPECT().RefreshCache(gomock.Any()).Return(2, nil)
	useCase.EXPECT().Evaluate(gomock.Any(), "flag", "user").Return(domain.Evaluation{}, service.ErrDependency)

	status, body := request(api, http.MethodPost, "/api/v1/internal/features", `{"name":"flag","description":"test","status":"open","whitelist":[]}`)
	assert.Equal(t, http.StatusCreated, status)
	assert.Contains(t, body, `"name":"flag"`)
	status, _ = request(api, http.MethodPost, "/api/v1/internal/features", `{`)
	assert.Equal(t, http.StatusBadRequest, status)
	status, body = request(api, http.MethodGet, "/api/v1/internal/features/missing", "")
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, body, "feature not found")
	status, _ = request(api, http.MethodDelete, "/api/v1/internal/features/flag", "")
	assert.Equal(t, http.StatusNoContent, status)
	status, body = request(api, http.MethodPost, "/api/v1/internal/cache/refresh", "")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, `"refreshed":2`)
	status, _ = request(api, http.MethodGet, "/api/v1/external/features/flag/users/user/evaluation", "")
	assert.Equal(t, http.StatusServiceUnavailable, status)
}

func TestFeatureRoutesSuccessfulReadsAndUpdate(t *testing.T) {
	controller := gomock.NewController(t)
	useCase := mocks.NewMockFeatureUseCase(controller)
	api := router.New(handler.NewFeatureHandler(useCase))
	feature := domain.Feature{Name: "flag", Status: domain.StatusOpen}
	useCase.EXPECT().List(gomock.Any()).Return([]domain.Feature{feature}, nil)
	useCase.EXPECT().Get(gomock.Any(), "flag").Return(feature, nil)
	useCase.EXPECT().Update(gomock.Any(), "flag", service.WriteFeature{Status: domain.StatusClosed, Whitelist: []string{}}).Return(domain.Feature{Name: "flag", Status: domain.StatusClosed}, nil)
	useCase.EXPECT().Evaluate(gomock.Any(), "flag", "user").Return(domain.Evaluation{FeatureName: "flag", UserID: "user", Enabled: true, Status: domain.StatusOpen}, nil)
	useCase.EXPECT().Health(gomock.Any()).Return(nil)

	status, _ := request(api, http.MethodGet, "/api/v1/internal/features", "")
	assert.Equal(t, http.StatusOK, status)
	status, _ = request(api, http.MethodGet, "/api/v1/internal/features/flag", "")
	assert.Equal(t, http.StatusOK, status)
	status, _ = request(api, http.MethodPut, "/api/v1/internal/features/flag", `{"description":"","status":"closed","whitelist":[]}`)
	assert.Equal(t, http.StatusOK, status)
	status, _ = request(api, http.MethodGet, "/api/v1/external/features/flag/users/user/evaluation", "")
	assert.Equal(t, http.StatusOK, status)
	status, _ = request(api, http.MethodGet, "/healthz", "")
	assert.Equal(t, http.StatusOK, status)
}

func TestHandlerMapsValidationConflictAndUnexpectedErrors(t *testing.T) {
	controller := gomock.NewController(t)
	useCase := mocks.NewMockFeatureUseCase(controller)
	api := router.New(handler.NewFeatureHandler(useCase))
	useCase.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.Feature{}, service.ErrValidation)
	useCase.EXPECT().List(gomock.Any()).Return(nil, repository.ErrConflict)
	useCase.EXPECT().Health(gomock.Any()).Return(errors.New("unexpected"))

	status, _ := request(api, http.MethodPost, "/api/v1/internal/features", `{"name":"flag","description":"","status":"open","whitelist":[]}`)
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = request(api, http.MethodGet, "/api/v1/internal/features", "")
	assert.Equal(t, http.StatusConflict, status)
	status, _ = request(api, http.MethodGet, "/healthz", "")
	assert.Equal(t, http.StatusInternalServerError, status)
}

func request(api http.Handler, method, path, body string) (int, string) {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	return response.Code, response.Body.String()
}

func TestRouterRejectsUnknownRoute(t *testing.T) {
	controller := gomock.NewController(t)
	api := router.New(handler.NewFeatureHandler(mocks.NewMockFeatureUseCase(controller)))
	status, _ := request(api, http.MethodGet, "/unknown", "")
	require.Equal(t, http.StatusNotFound, status)
}
