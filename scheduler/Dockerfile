FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /scheduler .

FROM scratch
COPY --from=builder /scheduler /scheduler
ENTRYPOINT ["/scheduler"]
