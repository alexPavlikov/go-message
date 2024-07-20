package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alexPavlikov/go-message/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cancel context.CancelFunc, cfg config.Config) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Postrges.User, cfg.Postrges.Password, cfg.Postrges.Server.Path, cfg.Postrges.Server.Port, cfg.Postrges.DB))
	if err != nil {
		slog.Error("failed connect to postgres", "error", err)
		return nil, err
	}

	defer cancel()

	slog.Info("connection to db complited")

	return db, nil
}
