# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build all services
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api/main.go && \
    CGO_ENABLED=0 go build -o /app/listener ./cmd/listener/main.go && \
    CGO_ENABLED=0 go build -o /app/scheduler ./cmd/scheduler/main.go && \
    CGO_ENABLED=0 go build -o /app/sender ./cmd/sender/main.go

# Final stage
FROM alpine:3.18

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binaries
COPY --from=builder /app/api /app/api
COPY --from=builder /app/listener /app/listener
COPY --from=builder /app/scheduler /app/scheduler
COPY --from=builder /app/sender /app/sender

# Default command (can be overridden in docker-compose)
CMD ["/app/api"]