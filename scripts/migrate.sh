#!/bin/bash
# ============================================================
# FILE: scripts/migrate.sh
# WHAT IT IS:     Run database migrations
# NOTE:           GORM AutoMigrate runs automatically on server start.
#                 This script runs the reference SQL files manually
#                 if needed for inspection or manual DB setup.
# HOW TO RUN:     bash scripts/migrate.sh
# ============================================================

set -e
source "$(dirname "$0")/../backend/.env"

DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
echo "Running migrations against: $DB_URL"

for f in "$(dirname "$0")/../backend/migrations/"*.sql; do
  echo "  Running: $(basename $f)"
  psql "$DB_URL" -f "$f"
done

echo "✅ Migrations complete"
