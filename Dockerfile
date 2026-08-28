# Build stage
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

# TARGETARCH is automatically set by Docker
ARG TARGETARCH

# Create a non-root user
RUN adduser -D -u 10001 appuser

# Ensure modules are on and build is reproducible
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=$TARGETARCH

WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of your source code
COPY . .

ARG Version=dev
# Build the Go binary (static binary for scratch image)
# -trimpath removes local filesystem paths from the binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w -X 'products/internal/system.Version=${Version}'" -o server .

# Final, minimal image
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/service-atlas/products" \
      org.opencontainers.image.description="Service Atlas Products microservice" \
      org.opencontainers.image.licenses="MIT"

# Copy CA certificates for TLS (e.g., Neo4j over TLS or outbound HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the passwd file to use the non-root user
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /app/server /server

# Use the non-root user
USER appuser

EXPOSE 8080

ENTRYPOINT ["/server"]