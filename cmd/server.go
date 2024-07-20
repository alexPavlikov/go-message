package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alexPavlikov/go-message/internal/config"
	postgres "github.com/alexPavlikov/go-message/internal/db"
	"github.com/alexPavlikov/go-message/internal/kafka"
	router "github.com/alexPavlikov/go-message/internal/server"
	"github.com/alexPavlikov/go-message/internal/server/locations"
	"github.com/alexPavlikov/go-message/internal/server/repository"
	"github.com/alexPavlikov/go-message/internal/server/service"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	_ "github.com/alexPavlikov/go-message/internal/migrations"
)

func Run() error {

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed run server - load config err: %w", err)
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(cfg.Timeout*time.Second))

	db, err := postgres.Connect(ctx, cancel, *cfg)
	if err != nil {
		return fmt.Errorf("failed run server - connect to postgres: %w", err)
	}

	conn := stdlib.OpenDBFromPool(db)

	if err := goose.Up(conn, "."); err != nil {
		return fmt.Errorf("failed to start migrations: %w", err)
	}

	producer, err := kafka.GetProducer(cfg.Kafka.ToString())
	if err != nil {
		return fmt.Errorf("failed run server - kafka get producer: %w", err)
	}

	consumer, close, err := kafka.GetConsumer(cfg.Kafka.ToString())
	if err != nil {
		return fmt.Errorf("failed run server - kafka get consumer: %w", err)
	}

	defer close()

	repo := repository.New(db, producer, cfg, consumer)
	service := service.New(repo)
	handler := locations.New(service)
	router := router.New(handler)
	srv := router.Build()

	if err := http.ListenAndServe(cfg.Server.ToString(), srv); err != nil {
		return fmt.Errorf("failed run server: %w", err)
	}

	return nil
}
