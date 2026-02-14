# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o config-service ./cmd/main.go

# Stage 2: Run
FROM alpine:3.18

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/config-service .

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./config-service"]
