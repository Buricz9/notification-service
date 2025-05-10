FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /scheduler ./cmd/scheduler/main.go

FROM scratch
COPY --from=builder /scheduler /app
ENTRYPOINT ["/app"]
