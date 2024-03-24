start:
	@echo "Starting all containers..."
	@docker-compose up --build -d
	@echo "All containers started."

stop:
	@echo "Stopping all containers..."
	@docker-compose down
	@echo "All containers stopped."

clean:
	@echo "Removing all containers..."
	@docker-compose down -v --rmi all --remove-orphans
	@echo "All containers removed."
