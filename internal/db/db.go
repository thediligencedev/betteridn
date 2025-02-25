package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Test the database connection
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return pool, nil
}
