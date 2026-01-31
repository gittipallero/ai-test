all: build-frontend run-backend

build-frontend:
	cd frontend && npm run build

run-backend:
	cd backend && DB_HOST=localhost DB_PORT=5434 DB_USER=postgres DB_PASSWORD=password DB_NAME=pacmangame DB_SSLMODE=disable go run .

dev-frontend:
	cd frontend && npm run dev
