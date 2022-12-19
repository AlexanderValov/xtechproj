
build:
	docker-compose build server

run:
	docker-compose up server

test:
	go test -v ./...
