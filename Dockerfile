# Build stage
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

# TARGETARCH is automatically set by Docker
ARG TARGETARCH

# Ensure modules are on and build is reproducible
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=$TARGETARCH

WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod ./
COPY go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of your source code
COPY . .

ARG Version=dev
RUN echo "Building version: ${Version}"
# Build the Go binary (static binary for scratch image)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w -X 'products/internal/system.Version=${Version}'" -o server .

# Final, minimal image
FROM scratch

# Copy CA certificates for TLS (e.g., Neo4j over TLS or outbound HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]