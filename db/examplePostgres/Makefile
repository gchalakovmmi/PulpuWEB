BINARY := examplePostgres
DB_CONTAINER := pg-test
PORT := 5432

.PHONY: run clean

run:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(DB_CONTAINER) || true
	@echo "Starting PostgreSQL container..."
	@docker run --rm --name $(DB_CONTAINER) -e POSTGRES_PASSWORD=postgres -p $(PORT):5432 -d postgres:17.6-alpine3.22
	@sleep 2 # Wait for DB to start
	@echo "Building and running application..."
	@go build -o $(BINARY) .
	@env POSTGRES_USER=postgres POSTGRES_PASSWORD=postgres POSTGRES_IP=localhost POSTGRES_PORT=$(PORT) POSTGRES_DB=postgres PORT=8080 ./$(BINARY)

clean:
	@docker stop $(DB_CONTAINER) || true
	@rm -f $(BINARY)
