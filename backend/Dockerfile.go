# Multi-stage build: Go binary + Python runtime for ML CLI
# Stage 1: Build Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Runtime with Python for ML CLI
FROM python:3.13-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install uv for Python dependency management
RUN pip install uv

# Silence ONNXRuntime CPU vendor warnings
ENV ORT_LOGGING_LEVEL=ERROR

# Copy Go binary from builder
COPY --from=builder /build/server /app/server

# Copy Python project files
COPY pyproject.toml uv.lock ./
COPY backend/requirements.txt ./backend/
COPY src/ ./src/
COPY scripts/ ./scripts/
COPY logger/ ./logger/

# Install Python dependencies
RUN uv pip install --system -r backend/requirements.txt

# Create required directories
RUN mkdir -p outputs logs

# Expose the API port
EXPOSE 8000

# Run the Go server
CMD ["/app/server"]
