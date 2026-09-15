package router

import (
	"feature-flag-manager/internal/http/handler"

	"github.com/gin-gonic/gin"
)

func New(featureHandler *handler.FeatureHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/healthz", featureHandler.Health)

	internal := router.Group("/api/v1/internal")
	internal.POST("/features", featureHandler.Create)
	internal.GET("/features", featureHandler.List)
	internal.GET("/features/:name", featureHandler.Get)
	internal.PUT("/features/:name", featureHandler.Update)
	internal.DELETE("/features/:name", featureHandler.Delete)
	internal.POST("/cache/refresh", featureHandler.RefreshCache)

	external := router.Group("/api/v1/external")
	external.GET("/features/:name/users/:userId/evaluation", featureHandler.Evaluate)
	return router
}
