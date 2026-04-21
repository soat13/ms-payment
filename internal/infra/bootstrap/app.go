package bootstrap

import (
	"context"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/payment/internal/application"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/application/ports/out"
	messagingHandler "github.com/soat13/payment/internal/infra/in/messaging/create_payment"
)

type (
	App struct {
		Repository     out.Repository
		TopicPublisher out.TopicPublisher
		QueuePublisher out.QueueSender

		// private
		consumer messaging.Consumer
	}
)

func NewApp(
	repository out.Repository,
	topicPublisher out.TopicPublisher,
	queueSender out.QueueSender,
	queueConsumer messaging.Consumer,
) *App {
	return &App{
		Repository:     repository,
		TopicPublisher: topicPublisher,
		QueuePublisher: queueSender,
		consumer:       queueConsumer,
	}
}

func (a *App) Start(ctx context.Context) {
	useCase := create_payment.NewCreatePaymentUseCase(a.Repository, a.TopicPublisher)
	handler := messagingHandler.Handler(useCase)

	a.consumer.Subscribe(application.PaymentRequestQueue, handler)

	go a.consumer.Listen(ctx)
}

func (a *App) Stop() {
	a.consumer.Stop()
}
