# Build stage
FROM golang:1.25.6-alpine as builder
WORKDIR /app
COPY mod.sum mod.lock ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
# Migrations
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080
CMD ["./server"]