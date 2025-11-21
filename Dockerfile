# Multi-stage Dockerfile for uptool
# Stage 1: Build the binary
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(cat internal/version/VERSION)" \
    -o /build/uptool \
    ./cmd/uptool

# Stage 2: Create minimal runtime image
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    curl \
    bash \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 uptool && \
    adduser -D -u 1000 -G uptool uptool

# Copy binary from builder
COPY --from=builder /build/uptool /usr/local/bin/uptool

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set working directory
WORKDIR /workspace

# Change ownership
RUN chown -R uptool:uptool /workspace

# Switch to non-root user
USER uptool

# Set default command
ENTRYPOINT ["/usr/local/bin/uptool"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="uptool" \
      org.opencontainers.image.description="Universal manifest-first dependency updater" \
      org.opencontainers.image.vendor="santosr2" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/santosr2/uptool" \
      org.opencontainers.image.documentation="https://santosr2.github.io/uptool/"
