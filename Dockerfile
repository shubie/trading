# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# RUN CGO_ENABLED=0 GOOS=linux go build -o trading .
RUN CGO_ENABLED=0 GOOS=linux go build -o trading ./cmd/trading

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/trading .
COPY --from=builder /app/configs/config.yaml ./configs/

EXPOSE 8080 50057
CMD ["./trading"]