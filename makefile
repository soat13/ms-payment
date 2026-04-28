# Docker
up:
	docker compose up

up-d:
	docker compose up -d

down:
	docker compose down

# Go
tidy:
	docker compose exec payment go mod tidy
vendor:
	docker compose exec payment go mod vendor


# Tests
test:
	docker compose exec payment go test ./... -v --count=1
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