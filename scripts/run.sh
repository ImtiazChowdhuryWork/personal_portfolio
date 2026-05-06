#!/bin/bash
# ============================================================
# FILE: scripts/run.sh
# WHAT IT IS:     Start the Go backend server
# HOW TO RUN:     bash scripts/run.sh
# ============================================================

set -e
cd "$(dirname "$0")/../backend"
echo "🚀 Starting backend server..."
go run ./cmd/main.go
