package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const migrationTable = "agent_schema_migrations"

type (
	PostgresRepository struct {
		*Queries
		db *pgxpool.Pool
	}
)

func NewRepository(config *config.PostgresConfig) *PostgresRepository {
	ctx := context.Background()
	db, err := newPostgresDB(ctx, config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new postgres db connection")
	}

	if err = runMigrations(config); err != nil {
		log.Fatal().Err(err).Msg("Failed to run db migrations")
	}

	return &PostgresRepository{
		Queries: New(db),
		db:      db,
	}
}

func newPostgresDB(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	ConnConfig, err := pgxpool.ParseConfig(
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", cfg.Username, cfg.Password, cfg.Database, cfg.Host, cfg.Port, cfg.SSLMode))
	conn, err := pgxpool.NewWithConfig(ctx, ConnConfig)
	if err != nil {
		return nil, appError.ErrPostgres.WithError(err).WithMessage("Failed to create new postgres db connection").Err()
	}

	// ping db
	if err = conn.Ping(ctx); err != nil {
		return nil, appError.ErrPostgres.WithError(err).WithMessage("Failed to ping db").Err()
	}

	return conn, nil

}

func runMigrations(cfg *config.PostgresConfig) error {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", cfg.Username, cfg.Password, cfg.Database, cfg.Host, cfg.Port, cfg.SSLMode))
	if err != nil {
		return err
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close db connection after running migrations")
		}
	}()
	driver, err := pg.WithInstance(db, &pg.Config{
		MigrationsTable: migrationTable,
		DatabaseName:    cfg.Database,
	})
	if err != nil {
		return appError.ErrPostgres.WithError(err).WithMessage("Failed to create new postgres driver").Err()
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationPath),
		cfg.Database,
		driver,
	)

	if err != nil {
		return appError.ErrPostgres.WithError(err).WithMessage("Failed to create new postgres db connection").Err()
	}

	if err = m.Up(); err != nil {
		if !errors.Is(migrate.ErrNoChange, err) {
			return appError.ErrPostgres.WithError(err).WithMessage("Failed to run migrations").Err()
		}
	}
	return nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}
