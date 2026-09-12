build:
	docker compose build

run:
	docker compose up

generate-docs:
	swag init -d ./,cmd/server,internal/handlers --parseInternal -g cmd/server/main.go

recreate:
	docker compose down
	docker compose up --build --force-recreate