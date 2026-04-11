# Docker
docker-up:
	docker compose up -d
docker-down:
	docker compose down

# Tests
test:
	go test ./... -v --count=1
mock:
	 mockgen -package=mock -source=internal/application/ports/out/repository.go -destination=internal/application/ports/out/mock/repository.go

# Go
mod:
	go mod tidy
vendor:
	go mod vendor

