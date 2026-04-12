# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Copy the rest of your source code
COPY . .

# Run tests
RUN go test ./...


# Build the Go binary (static binary for scratch image)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Final, minimal image
FROM scratch

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]