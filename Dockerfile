FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o jwt_refresher -ldflags="-s -w" .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs tzdata

# Set timezone
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/jwt_refresher /app/jwt_refresher

# Make binary executable
RUN chmod +x /app/jwt_refresher

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 3007

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3007/ || exit 1

# Run the application
CMD ["/app/jwt_refresher"]
