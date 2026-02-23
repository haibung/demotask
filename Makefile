.PHONY: help build up down restart logs ps migrate-up migrate-down migrate-create backend-shell db-shell tunnel tunnel-down

help:
	@echo "Available commands:"
	@echo "  make build           - Build backend image"
	@echo "  make up              - Start backend + db"
	@echo "  make down            - Stop all services"
	@echo "  make restart         - Restart backend + db"
	@echo "  make logs            - Follow logs"
	@echo "  make ps              - Show container status"
	@echo "  make migrate-up      - Run migration up"
	@echo "  make migrate-down    - Run migration down"
	@echo "  make migrate-create name=table_name - Create migration"
	@echo "  make backend-shell   - Open backend container shell"
	@echo "  make db-shell        - Open PostgreSQL shell"
	@echo "  make tunnel          - Start ngrok tunnel"
	@echo "  make tunnel-down     - Stop ngrok tunnel"

build:
	docker compose build backend

up:
	docker compose up -d --build

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f --tail=200

ps:
	docker compose ps

migrate-up:
	docker compose exec backend /app/app migrate up

migrate-down:
	docker compose exec backend /app/app migrate down

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=table_name" && exit 1)
	docker compose exec backend /app/app migrate create $(name)

backend-shell:
	@echo "Backend shell is disabled because the image uses distroless (no shell available)."

db-shell:
	docker compose exec db psql -U $${POSTGRES_USER:-demotask} -d $${POSTGRES_DB:-demotask}

tunnel:
	docker compose --profile tunnel up -d ngrok

tunnel-down:
	docker compose --profile tunnel stop ngrok
