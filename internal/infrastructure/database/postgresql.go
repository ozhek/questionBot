package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

var pgdb *pgxpool.Pool

// InitializePostgres initializes the PostgreSQL database connection.
func InitializePostgres(cfg PostgresConfig) {
	dsn := "postgres://" + cfg.User + ":" + cfg.Password + "@" + cfg.Host + ":" + cfg.Port + "/" + cfg.DBName + "?sslmode=" + cfg.SSLMode

	if dsn == "" {
		log.Fatal("PostgreSQL DSN is empty")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("unable to parse DSN: %v", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime

	pgdb, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("failed to create PostgreSQL connection pool: %v", err)
	}

	if err = pgdb.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping PostgreSQL: %v", err)
	}

	log.Println("Connected to PostgreSQL")
}

// GetPostgresDB returns the PostgreSQL connection pool.
func GetPostgresDB() *pgxpool.Pool {
	return pgdb
}
