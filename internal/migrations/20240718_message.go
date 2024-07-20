package migrations

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upMessage, downMigration)
}

func upMessage(tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS message (
	"id" serial PRIMARY KEY NOT NULL,
	"chat_id" integer NOT NULL,
	"user_id" integer NOT NULL,
	"text" text NOT NULL,
	"time" timestamp_with_timezone NOT NULL,
	"db" boolean NOT NULL,
	"kafka" boolean NOT NULL,
	"deleted" boolean NOT NULL);`

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("migrations failed to create table message: %w", err)
	}

	return nil
}

func downMigration(tx *sql.Tx) error {
	query := `DROP TABLE message;`

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("migrations failed to drop table message: %w", err)
	}

	return nil
}
