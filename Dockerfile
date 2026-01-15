FROM alpine:latest

# Install ca-certificates and sqlite for runtime
RUN apk --no-cache add ca-certificates sqlite-libs tzdata

# Set timezone to Asia/Shanghai (optional, adjust as needed)
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy the precompiled Linux binary
COPY jwt_refresher /app/jwt_refresher

# Make binary executable
RUN chmod +x /app/jwt_refresher

# Copy embedded web files (they're embedded in the binary, but keep structure for reference)
# The web files are actually embedded via go:embed, so this is just for documentation

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 3007

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3007/ || exit 1

# Run the application
CMD ["/app/jwt_refresher"]
