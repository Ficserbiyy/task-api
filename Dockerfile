# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files first for Docker layer caching.
COPY go.mod go.sum ./

RUN go mod download

# Copy source code.
COPY . .

# Build a static Linux binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /bin/task-api \
    ./cmd/server


FROM alpine:3.22

WORKDIR /app

# CA certificates are useful for HTTPS requests.
RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/task-api /app/bin/task-api

EXPOSE 8000

ENTRYPOINT ["/app/bin/task-api"]