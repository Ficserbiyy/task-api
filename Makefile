run:
	go run .

build:
	go build -o ./bin/task-api

run-binary:
	./bin/task-api

run-postgres:
	docker run --name my-postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres

test:
	go test ./...