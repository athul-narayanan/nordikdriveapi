# ----------------------
# Builder stage
# ----------------------
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install git (openssl not needed if you’re not building HTTPS inside)
RUN apk add --no-cache git

# Copy Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o nordik-drive-api ./cmd/server

# ----------------------
# Final stage
# ----------------------
FROM alpine:latest
WORKDIR /app

# Install CA certificates (for outbound HTTPS requests to other services)
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /app/nordik-drive-api .

# Expose Cloud Run HTTP port
EXPOSE 8080

# Run the server (Cloud Run expects it to bind to $PORT or 8080)
CMD ["./nordik-drive-api"]
