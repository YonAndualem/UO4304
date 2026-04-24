.PHONY: dev_up dev_down dev_logs dev_ps dev_seed dev_reset dev_rebuild

# Start the full local stack (postgres + minio + app + frontend)
dev_up:
	docker compose up --build -d

# Stop and remove running containers (keeps volumes)
dev_down:
	docker compose down

# Rebuild and restart everything from scratch
dev_rebuild:
	docker compose down
	docker compose up --build -d

# Follow combined logs from all services
dev_logs:
	docker compose logs -f

# Show running service status
dev_ps:
	docker compose ps

# Seed demo data once
dev_seed:
	docker compose run --rm seed

# Reset non-demo data
dev_reset:
	docker compose run --rm reset
