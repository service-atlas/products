# use PowerShell instead of sh on Windows:
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

set quiet := true

# Run tests (short)
[default]
test:
  echo "Running tests"
  go test --short -v ./...

# Run all tests (including long ones)
test-full:
  echo "Running all tests"
  go test -v ./...

# Run tests with short output and coverage
test-cover:
  echo "Running tests with coverage"
  go test --short -v ./... -covermode=count -coverprofile={{invocation_directory()}}/coverage.out
  go tool cover -func {{invocation_directory()}}/coverage.out

# Run all tests with coverage
test-full-cover:
  echo "Running all tests with coverage"
  go test -v ./... -covermode=count -coverprofile={{invocation_directory()}}/coverage.out
  go tool cover -func {{invocation_directory()}}/coverage.out

# Generates schema from liquibase then converts it to SQLC golang code
generate_schema: liquibase_update_sql sqlc_generate

# Generates schema.sql from liquibase based on the sqlc context
[working-directory: "migrations"]
liquibase_update_sql:
  liquibase update-sql --context-filter="sqlc" --output-file=../schema.sql

# Generates SQLC golang code from schema.sql
sqlc_generate:
  echo "Generating SQL code into go"
  sqlc generate
  echo "Go code generated"

# Run liquibase update
[working-directory: "migrations"]
lbupdate:
  liquibase update

# Run database in docker
up:
  docker-compose up -d

# Stop database
stop:
  docker-compose stop