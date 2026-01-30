#!/bin/bash

# Start PostgreSQL using Docker Compose
echo "Starting PostgreSQL..."
docker compose up -d

# Wait for Postgres to be ready (naive check)
echo "Waiting for DB..."
sleep 5

# Export environment variables for the backend
export DB_HOST=localhost
export DB_PORT=5434
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=pacmangame
export DB_SSLMODE=disable

echo "Environment variables set."
echo ""
echo "To run the backend:"
echo "  cd backend && DB_HOST=localhost DB_PORT=5434 DB_USER=postgres DB_PASSWORD=password DB_NAME=pacmangame DB_SSLMODE=disable go run main.go"
echo ""
echo "To run the frontend:"
echo "  cd frontend && npm run dev"
