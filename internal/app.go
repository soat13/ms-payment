package app

import (
	"context"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/application/ports/out"
	messagingHandler "github.com/soat13/payment/internal/infra/in/messaging/create_payment"
)

const PaymentRequestQueue = "payment_request"

type (
	Container interface {
		GetBroker(ctx context.Context) (messaging.Broker, error)
		GetRepository(ctx context.Context) (out.Repository, error)
	}

	App struct {
		Repository out.Repository
		Broker     messaging.Broker
	}
)

func New(ctx context.Context, container Container) (*App, error) {

	repository, err := container.GetRepository(ctx)
	if err != nil {
		return nil, err
	}

	broker, err := container.GetBroker(ctx)
	if err != nil {
		return nil, err
	}

	return &App{
		Broker:     broker,
		Repository: repository,
	}, nil
}

func (a *App) Start(ctx context.Context) {
	useCase := create_payment.NewCreatePaymentUseCase(a.Repository)
	handler := messagingHandler.Handler(useCase)

	a.Broker.Subscribe(PaymentRequestQueue, handler)

	go a.Broker.Listen(ctx)
}

func (a *App) Stop() {
	a.Broker.Stop()
}
