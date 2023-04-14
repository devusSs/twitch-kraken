package postgres

import (
	"database/sql"
	"fmt"

	"github.com/devusSs/twitch-kraken/internal/config"
	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/database/postgres/statements"
	_ "github.com/lib/pq"
)

// Internal Postgres structure which executes database.Service layer functions.
type psql struct {
	db *sql.DB
}

// Inits a new Postgres connection and returns database.Service layer.
func New(cfg *config.Config) (database.Service, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password,
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database)

	db, err := sql.Open("postgres", dsn)

	return &psql{db}, err
}

// Test database connection.
func (p *psql) Ping() error {
	return p.db.Ping()
}

// Closes the database connection.
func (p *psql) Close() error {
	return p.db.Close()
}

// Migrates models / creates tables (check statements.go) on database.
func (p *psql) Migrate() error {
	_, err := p.db.Exec(statements.CreateGateKeeperSettingsStore)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(statements.CreateUsersTable)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(statements.CreateCommandsTable)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(statements.CreateAuthEventsTable)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(statements.CreateMessageEventsTable)
	if err != nil {
		return err
	}

	return nil
}
