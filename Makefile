all: build-frontend run-backend

build-frontend:
	cd frontend && npm run build

# For local development, create a .env file from .env.example
# and source it before running, or use: make run-backend-dev
run-backend:
	cd backend && go run .

# Convenience target for local development with default settings
# NOTE: For production, use environment variables instead
run-backend-dev:
	@echo "Starting backend with development settings..."
	@echo "For production, set environment variables properly."
	cd backend && \
		DB_HOST=localhost \
		DB_PORT=5434 \
		DB_USER=postgres \
		DB_PASSWORD=$${DB_PASSWORD:-password} \
		DB_NAME=pacmangame \
		DB_SSLMODE=disable \
		ALLOWED_ORIGINS=http://localhost:6060,http://localhost:5173 \
		go run .

dev-frontend:
	cd frontend && npm run dev
