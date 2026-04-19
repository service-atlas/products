# use PowerShell instead of sh:
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

set quiet := true

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


# Run tests with short output
[default]
test:
  echo "Running tests"
  go test -v --short ./...

# Run database in docker
db_up:
  docker-compose up -d

# Stop database
db_stop:
  docker-compose stop