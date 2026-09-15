package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"feature-flag-manager/internal/http/handler"
	"feature-flag-manager/internal/http/handler/mocks"
	"feature-flag-manager/internal/http/router"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewRegistersHealthRoute(t *testing.T) {
	controller := gomock.NewController(t)
	useCase := mocks.NewMockFeatureUseCase(controller)
	useCase.EXPECT().Health(gomock.Any()).Return(nil)
	response := httptest.NewRecorder()
	router.New(handler.NewFeatureHandler(useCase)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	assert.Equal(t, http.StatusOK, response.Code)
}
