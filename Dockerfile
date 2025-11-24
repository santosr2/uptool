# Dockerfile for uptool
# Multi-stage build for optimized image size

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary for target architecture
ARG TARGETARCH
ARG TARGETOS=linux
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o /usr/local/bin/uptool ./cmd/uptool

# Runtime stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    curl \
    bash \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 uptool && \
    adduser -D -u 1000 -G uptool uptool

# Copy binary from builder
COPY --from=builder --chown=uptool:uptool /usr/local/bin/uptool /usr/local/bin/uptool

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
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/santosr2/uptool" \
      org.opencontainers.image.documentation="https://santosr2.github.io/uptool/"
