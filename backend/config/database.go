// ============================================================
// FILE: config/database.go
// WHAT IT IS:     PostgreSQL database connection setup using GORM
// WHY IT EXISTS:  Separates database wiring from business logic —
//                 the rest of the app just uses the *gorm.DB object
// ENDPOINTS:      N/A — internal setup only
// DEPENDS ON:     config.go, gorm.io/gorm, gorm.io/driver/postgres
// IF REMOVED:     No database connection — all data operations fail
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package config

import (
	"fmt"
	"log"

	// GORM is the ORM (Object-Relational Mapper) that lets us work
	// with database rows as Go structs instead of raw SQL queries
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/**
 * FUNCTION: ConnectDB
 * WHAT IT DOES:   Opens a connection to PostgreSQL using GORM,
 *                 configures the connection pool, and returns
 *                 a ready-to-use *gorm.DB instance.
 *                 Also enables SQL logging in development mode
 *                 so you can see every query in the terminal.
 * WHY IT EXISTS:  All database access goes through this single
 *                 connection — sharing one pool is more efficient
 *                 than opening a new connection per request
 * WHERE CALLED:   cmd/main.go → at startup, before routes are set up
 * PARAMETERS:     @param {*Config} cfg - app config with DB credentials
 * RETURNS:        @returns {*gorm.DB} - active database connection
 * IF CHANGED:     Any handler that receives *gorm.DB will be affected
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func ConnectDB(cfg *Config) *gorm.DB {
	// ─── Step 1: Choose log level based on environment ────────
	// In development we want to see every SQL query printed to
	// the terminal so we can debug what GORM is actually doing.
	// In production we only log slow queries and errors to avoid
	// flooding production logs with routine SELECT statements.
	logLevel := logger.Info // development: show all queries
	if cfg.AppEnv == "production" {
		logLevel = logger.Warn // production: only warn/error
	}

	// ─── Step 2: Open the database connection ─────────────────
	// gorm.Open connects to PostgreSQL using the DSN string we
	// built in config.go. It does NOT run any queries yet —
	// it just establishes the connection pool.
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		// If we can't connect to the database the server cannot run.
		// log.Fatal prints the error and exits the process immediately.
		log.Fatal("Failed to connect to database: ", err)
	}

	// ─── Step 3: Get the underlying sql.DB to configure the pool ─
	// GORM wraps the standard library's sql.DB. We access it here
	// to set connection pool limits, preventing the app from
	// opening hundreds of connections under load.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB: ", err)
	}

	// Maximum number of connections sitting idle in the pool.
	// Idle connections are kept open so they can be reused quickly.
	sqlDB.SetMaxIdleConns(10)

	// Maximum number of connections open at the same time.
	// If all 100 are in use, new requests wait for one to free up.
	sqlDB.SetMaxOpenConns(100)

	// ─── Step 4: Confirm the connection is alive ──────────────
	// Ping sends a lightweight query to make sure the DB is
	// actually reachable — not just that GORM initialized without error
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Database ping failed — is PostgreSQL running? Error: ", err)
	}

	fmt.Println("✅ Database connected successfully")
	return db
}
