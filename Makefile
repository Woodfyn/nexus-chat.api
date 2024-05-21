.PHONY:
.SILENT:

init:
	go mod download

gorun:
	go run cmd/main.go

build-prod:
	docker build -f docker/prod/Dockerfile -t chat-api-prod .

build-debug:
	docker build -f docker/debug/Dockerfile -t chat-api-debug .

build:
	make build-prod
	make build-debug

run:
	docker-compose up -d --build

stop:
	docker-compose down

swagger:
	swag init -g cmd/main.go