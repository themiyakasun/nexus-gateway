package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(connStr string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5*time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully!")
	return db, nil
}

func AutoMigrate(db *sql.DB) error {
	query := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS organizations (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(255) NOT NULL,
				slug VARCHAR(100) UNIQUE NOT NULL,
				is_active BOOLEAN NOT NULL DEFAULT true,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS api_keys (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				key_hash VARCHAR(64) UNIQUE NOT NULL,
				key_prefix VARCHAR(16) NOT NULL,
				name VARCHAR(100) NOT NULL,
				is_active BOOLEAN NOT NULL DEFAULT true,
				rate_limit_rpm INT NOT NULL DEFAULT 60,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				expires_at TIMESTAMPTZ
	);

	CREATE TABLE IF NOT EXISTS services (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(255) NOT NULL,
				url VARCHAR(500) NOT NULL,
				health_check_path VARCHAR(255) NOT NULL DEFAULT '/health',
				is_active BOOLEAN NOT NULL DEFAULT true,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS routes (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
				path_prefix VARCHAR(255) NOT NULL UNIQUE,
				strip_prefix BOOLEAN NOT NULL DEFAULT false,
				auth_required BOOLEAN NOT NULL DEFAULT true,
				rate_limit_override_rpm INT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to run AutoMigrate: %w", err)
	}

	log.Println("Database migration completed: organizations, api_keys, services, routes verified.")
	return nil
}