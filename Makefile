include .env
export

# Start all services (PostgreSQL, migrations, and Go application)
all:
	docker-compose up -d --build

# Start PostgreSQL and run migrations as a one-off task
setup:
	docker-compose up -d postgres
	sleep 10 # Give time for the database to be ready
	docker-compose run --rm migrate

# Stop and remove all services, networks, and volumes
teardown:
	docker-compose down -v

# Remove data with confirmation
removedata:
	@echo "WARNING: This will permanently delete the database data!"
	@read -p "Are you sure you want to continue? [y/N]: " confirm && [ $$confirm = y ] || [ $$confirm = Y ] || (echo "Aborted." && exit 1)
	sudo rm -rf ./data

# Clean up Docker resources that are not in use
cleanup:
	docker system prune -af --volumes


# Run migrations up
migration_up:
	docker-compose run --rm migrate -path=/migrations -database "${DATABASE_URL}" up

# Run migrations down (reversing migrations)
migration_down:
	docker-compose run --rm migrate -path=/migrations -database "${DATABASE_URL}" down

# Fix a specific version in migrations
migration_fix:
	docker-compose run --rm migrate -path=/migrations -database "${DATABASE_URL}" force 1

# Run tests
test:
	go test -v -cover ./...

.PHONY: all setup teardown removedata cleanup migration_down migration_fix test
