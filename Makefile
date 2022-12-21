
build:
	docker-compose build server

run:
	docker-compose up server

test:
	go test -v -count=1 ./...

test100:
	go test -v -count=100 ./...

race:
	go test -v -race -count=1 ./...

.PHONY: gen-repo
gen-repo:
	mockgen -source=internal/repository/repository.go \
	-destination=internal/repository/mocks/mock_repository.go

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out
