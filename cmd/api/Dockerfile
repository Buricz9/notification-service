# Dockerfile do budowy kontenera dla aplikacji API
# --- etap 1: build ---
FROM golang:1.24-alpine AS builder

# ustaw katalog roboczy w kontenerze
WORKDIR /app

# cache pobierania modułów
COPY go.mod ./
RUN go mod download

# skopiuj cały projekt i skompiluj tylko api
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin/api ./cmd/api

# --- etap 2: final ---
FROM alpine:latest

# tylko potrzebne pliki czasowe i certy (jeśli trzeba TLS)
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/bin/api .

# port zgodny z docker-compose.yml
EXPOSE 8080

# domyślna komenda
CMD ["./api"]
