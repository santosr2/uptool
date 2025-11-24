# Dockerfile for uptool
# Uses pre-built binaries from release artifacts

# ARG for target architecture (set by docker buildx)
ARG TARGETARCH

# Runtime image
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

# Copy pre-built binary (provided via build context)
# The binary should be at ./dist/uptool-linux-${TARGETARCH}/uptool
ARG TARGETARCH
COPY dist/uptool-linux-${TARGETARCH}/uptool /usr/local/bin/uptool
RUN chmod +x /usr/local/bin/uptool

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
