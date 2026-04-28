# Docker
docker-up:
	docker compose up

docker-up-d:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f payment

docker-test:
	docker compose run --rm payment go test ./...

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
	 mockgen -package=mock -source=internal/application/ports/out/topic_publisher.go -destination=internal/application/ports/out/mock/topic_publisher.go
	 mockgen -package=mock -source=internal/application/ports/out/payment_provider.go -destination=internal/application/ports/out/mock/payment_provider.go
	 mockgen -package=mock -source=internal/infra/out/providers/mercado_pago/client.go -destination=internal/infra/out/providers/mercado_pago/mock/client.go

# Localstack
aws-sqs-list:
	docker exec -it localstack-main awslocal sqs list-queues
aws-sns-list:
	docker exec -it localstack-main awslocal sns list-subscriptions