package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"feature-flag-manager/internal/cache/redis"
	"feature-flag-manager/internal/config"
	"feature-flag-manager/internal/http/handler"
	"feature-flag-manager/internal/http/router"
	"feature-flag-manager/internal/repository/postgres"
	"feature-flag-manager/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	configuration, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	applicationContext, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(applicationContext, configuration.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	options, err := goredis.ParseURL(configuration.RedisURL)
	if err != nil {
		log.Fatalf("parse REDIS_URL: %v", err)
	}
	redisClient := goredis.NewClient(options)
	defer redisClient.Close()

	featureRepository := postgres.NewFeatureRepository(pool)
	featureCache := redis.NewFeatureCache(redisClient)
	featureService := service.NewFeatureService(featureRepository, featureCache)
	server := &http.Server{Addr: configuration.HTTPAddr, Handler: router.New(handler.NewFeatureHandler(featureService))}

	go func() {
		log.Printf("HTTP server listening on %s", configuration.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve HTTP: %v", err)
		}
	}()

	<-applicationContext.Done()
	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("shutdown HTTP server: %v", err)
	}
}
