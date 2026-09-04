# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod first (and go.sum if it existed)
COPY go.mod ./
# RUN go mod download # Not needed if no external deps

# Copy source
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /simple-api main.go

# Run stage
FROM alpine:latest

# Security: Run as non-root
RUN adduser -D wallarm
USER wallarm

WORKDIR /

COPY --from=builder /simple-api /simple-api

EXPOSE 8080

CMD ["/simple-api"]
