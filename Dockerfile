# Stage 1: Build the Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
# Copy package.json and lock files
COPY frontend/package*.json ./
# Install dependencies
RUN npm ci
# Copy source code
COPY frontend/ .
# Build the frontend
RUN npm run build

# Stage 2: Build the Backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app/backend
# Copy go.mod and go.sum for checksum verification
COPY backend/go.mod backend/go.sum ./
# Download dependencies (if any)
RUN go mod download
# Copy source code
COPY backend/ .
# Build the Go binary
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o pacman-server .

# Stage 3: Final Image
FROM alpine:latest
WORKDIR /root/

# Copy the binary from the backend builder
COPY --from=backend-builder /app/backend/pacman-server .

# Copy the frontend build artifacts to the location expected by the backend
# The backend expects "frontend/dist" relative to its execution path or root
# modifying backend logic might be needed if paths differ, 
# but based on main.go: 
# "rootDir = ." and "filepath.Join(rootDir, "frontend/dist")"
# So if we run ./pacman-server, we need a directory "frontend/dist" next to it.
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Expose the port
EXPOSE 6060

# Command to run the executable
CMD ["./pacman-server"]
