package bootstrap

import (
	"context"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/payment/internal/application/create_link"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/infra"
	createLinkPaymentHandler "github.com/soat13/payment/internal/infra/in/messaging/create_link"
	createPaymentHandler "github.com/soat13/payment/internal/infra/in/messaging/create_payment"
)

type (
	App struct {
		container *Container
		consumer  messaging.Consumer
	}
)

func NewApp(container *Container) *App {
	return &App{
		container: container,
		consumer:  container.Consumer,
	}
}

func (a *App) Start(ctx context.Context) {

	a.consumer.Subscribe(
		infra.PaymentRequestQueue,
		createPaymentHandler.Handler(a.getCreatePaymentUseCase()),
	)

	a.consumer.Subscribe(
		infra.PaymentLinkRequest,
		createLinkPaymentHandler.Handler(a.getCreateLinkUseCase()),
	)

	go a.container.Consumer.Listen(ctx)
}

func (a *App) Stop() {
	a.container.Consumer.Stop()
}

func (a *App) getCreatePaymentUseCase() *create_payment.CreatePaymentUseCase {
	return create_payment.NewCreatePaymentUseCase(
		a.container.Repository,
		a.container.TopicPublisher,
	)
}

func (a *App) getCreateLinkUseCase() *create_link.CreateLinkUseCase {
	return create_link.NewCreateLinkUseCase(
		a.container.Repository,
		a.container.TopicPublisher,
		a.container.PaymentProvider,
	)
}
