
build:
	docker-compose build server

run:
	docker-compose up server

test:
	go test -v ./...

.PHONY: gen-repo
gen-repo:
	mockgen -source=internal/repository/repository.go \
	-destination=internal/repository/mocks/mock_repository.go

.PHONY: gen-serv
gen-serv:
	mockgen -source=internal/services/service.go \
	-destination=internal/services/mocks/mock_service.go