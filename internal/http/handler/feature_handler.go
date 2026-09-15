package handler

import (
	"context"
	"errors"
	"net/http"

	"feature-flag-manager/internal/domain"
	"feature-flag-manager/internal/repository"
	"feature-flag-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type FeatureHandler struct {
	service FeatureUseCase
}

type FeatureUseCase interface {
	Create(context.Context, service.WriteFeature) (domain.Feature, error)
	Get(context.Context, string) (domain.Feature, error)
	List(context.Context) ([]domain.Feature, error)
	Update(context.Context, string, service.WriteFeature) (domain.Feature, error)
	Delete(context.Context, string) error
	RefreshCache(context.Context) (int, error)
	Evaluate(context.Context, string, string) (domain.Evaluation, error)
	Health(context.Context) error
}

func NewFeatureHandler(service FeatureUseCase) *FeatureHandler {
	return &FeatureHandler{service: service}
}

type featureRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      domain.Status `json:"status"`
	Whitelist   []string      `json:"whitelist"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (handler *FeatureHandler) Create(context *gin.Context) {
	var request featureRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	feature, err := handler.service.Create(context.Request.Context(), writeFeature(request))
	if err != nil {
		handleError(context, err)
		return
	}
	context.JSON(http.StatusCreated, feature)
}

func (handler *FeatureHandler) List(context *gin.Context) {
	features, err := handler.service.List(context.Request.Context())
	if err != nil {
		handleError(context, err)
		return
	}
	context.JSON(http.StatusOK, features)
}

func (handler *FeatureHandler) Get(context *gin.Context) {
	feature, err := handler.service.Get(context.Request.Context(), context.Param("name"))
	if err != nil {
		handleError(context, err)
		return
	}
	context.JSON(http.StatusOK, feature)
}

func (handler *FeatureHandler) Update(context *gin.Context) {
	var request featureRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	feature, err := handler.service.Update(context.Request.Context(), context.Param("name"), writeFeature(request))
	if err != nil {
		handleError(context, err)
		return
	}
	context.JSON(http.StatusOK, feature)
}

func (handler *FeatureHandler) Delete(context *gin.Context) {
	if err := handler.service.Delete(context.Request.Context(), context.Param("name")); err != nil {
		handleError(context, err)
		return
	}
	context.Status(http.StatusNoContent)
}

func (handler *FeatureHandler) RefreshCache(context *gin.Context) {
	refreshed, err := handler.service.RefreshCache(context.Request.Context())
	if err != nil {
		handleError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"refreshed": refreshed})
}

func (handler *FeatureHandler) Evaluate(context *gin.Context) {
	evaluation, err := handler.service.Evaluate(context.Request.Context(), context.Param("name"), context.Param("userId"))
	if err != nil {
		handleError(context, err)
		return
	}
	context.JSON(http.StatusOK, evaluation)
}

func (handler *FeatureHandler) Health(context *gin.Context) {
	if err := handler.service.Health(context.Request.Context()); err != nil {
		handleError(context, err)
		return
	}
	context.Status(http.StatusOK)
}

func writeFeature(request featureRequest) service.WriteFeature {
	return service.WriteFeature{Name: request.Name, Description: request.Description, Status: request.Status, Whitelist: request.Whitelist}
}

func handleError(context *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		context.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, repository.ErrNotFound):
		context.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, repository.ErrConflict):
		context.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrDependency):
		context.JSON(http.StatusServiceUnavailable, errorResponse{Error: "dependency unavailable"})
	default:
		context.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
