all: build-frontend run-backend

build-frontend:
	cd frontend && npm run build

run-backend:
	cd backend && go run main.go

dev-frontend:
	cd frontend && npm run dev
