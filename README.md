# Chi Boilerplate

A lightweight, modular boilerplate for building RESTful APIs with Go and the Chi router framework. This project provides a clean architecture pattern with separation of concerns between handlers, services, and repositories.

Inspired by [Yappr](https://github.com/Melkeydev/yappr)

## Features

- Built with [Chi router](https://github.com/go-chi/chi) - a lightweight, idiomatic and composable router for building Go HTTP services
- Clean architecture with separation of concerns:
  - API handlers (presentation layer)
  - Services (business logic layer)
  - Repositories (data access layer)
- Middleware support:
  - Logging
  - Error recovery
  - Response compression
  - CORS configuration
- Environment variable configuration
- API testing with Bruno HTTP client
- Unit testing examples

## Project Structure

```
.
├── api/                  # API handlers (presentation layer)
│   ├── sample/           # Sample API endpoints
│   └── system/           # System API endpoints
├── HTTP_COLLECTION/      # Bruno HTTP client collection for API testing
├── repo/                 # Repositories (data access layer)
│   └── sample/           # Sample repository implementation
├── router/               # Router configuration
├── service/              # Services (business logic layer)
│   └── sample/           # Sample service implementation
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
└── main.go               # Application entry point
```

## API Endpoints

- `GET /api/time` - Returns the current server time
- `GET /api/sample/` - Returns a sample response
- `GET /api/sample/error` - Returns a sample error response

## Installation

### Prerequisites

- Go 1.24 or higher

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/chi-boilerplate.git
   cd chi-boilerplate
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable.

## Configuration

The application can be configured using environment variables:

- `PORT` - The port on which the server will listen (default: 8080)

## Testing

### Unit Tests

Run the unit tests with:

```bash
go test ./...
```

### API Testing

The project includes a Bruno HTTP client collection for testing the API endpoints. To use it:

1. Install [Bruno](https://www.usebruno.com/)
2. Open the HTTP_COLLECTION directory in Bruno
3. Use the provided requests to test the API endpoints

## Development

### Adding a New Endpoint

1. Create a new handler in the appropriate package under `api/`
2. Implement the necessary service logic in `service/`
3. If needed, add data access logic in `repo/`
4. Register the new endpoint in `router/router.go`

## License

This project is licensed under the MIT License - see the LICENSE file for details.
