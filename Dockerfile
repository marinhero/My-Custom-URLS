# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies (needed for SQLite CGO)
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o shortme .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/shortme .

# Expose port
EXPOSE 8080

# Run in server mode
CMD ["./shortme", "--server"]
