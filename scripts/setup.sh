#!/bin/bash
# ============================================================
# FILE: scripts/setup.sh
# WHAT IT IS:     Full project setup script
# HOW TO RUN:     bash scripts/setup.sh
# ============================================================

set -e
echo "🚀 Setting up Imtiaz Portfolio..."

# Copy .env.example to .env if not exists
if [ ! -f backend/.env ]; then
  cp backend/.env.example backend/.env
  echo "✓ .env created from .env.example — fill in your credentials"
fi

# Install Go dependencies
cd backend
go mod tidy
echo "✓ Go dependencies installed"

# Create uploads directory
mkdir -p uploads/images uploads/cv uploads/logos
echo "✓ Upload directories created"

cd ..
echo ""
echo "✅ Setup complete!"
echo "   1. Edit backend/.env with your PostgreSQL credentials"
echo "   2. Run: bash scripts/run.sh"
echo "   3. Open: frontend/public/index.html"
