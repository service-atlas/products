# use PowerShell instead of sh:
set shell := ["powershell.exe", "-c"]

set quiet := true

hello:
  Write-Host "Hello, world!"

generate_schema: liquibase_update_sql sqlc_generate

[working-directory: "migrations"]
liquibase_update_sql:
  liquibase update-sql --output-file=../schema.sql

sqlc_generate:
  Write-Host "Generating SQL code into go"
  sqlc generate
  Write-Host "Go code generated"

[default]
test:
  echo "Running tests"
  go test -v ./...

db_up:
  docker-compose up -d

db_stop:
  docker-compose stop