DB_FILE := knowledge.db
SQL_DIR := sql

.PHONY: init-db clean build

build:
	@echo "Building insight binary..."
	@go build -o insight .
	@echo "insight binary built."

init-db:
	@echo "Initializing database in $(DB_FILE)..."
	@cat $(SQL_DIR)/*.sql | sqlite3 $(DB_FILE)
	@echo "Database initialized."

clean:
	@echo "Cleaning up database..."
	@rm -f $(DB_FILE)
	@echo "Database cleaned."
