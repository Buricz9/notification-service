FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /sender .

FROM scratch
COPY --from=builder /sender /sender
ENTRYPOINT ["/sender"]
