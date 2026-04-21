# Docker
docker-up:
	docker compose up -d
docker-down:
	docker compose down

# Go
tidy:
	go mod tidy
vendor:
	go mod vendor


# Tests
test:
	go test ./... -v --count=1
mock:
	 mockgen -package=mock -source=internal/application/ports/out/repository.go -destination=internal/application/ports/out/mock/repository.go
	 mockgen -package=mock -source=internal/application/ports/out/publisher.go -destination=internal/application/ports/out/mock/publisher.go

# Localstack
aws-sqs-list:
	docker exec -it localstack-main awslocal sqs list-queues
