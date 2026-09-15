package handler

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mocks/handler_mocks.go -package=mocks feature-flag-manager/internal/http/handler FeatureUseCase
