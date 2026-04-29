package infra

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soat13/oficina-utils/pkg/db/ddb"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/oficina-utils/pkg/observability"
	"github.com/soat13/payment/internal/application/create_link"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/application/process_payment_status"
	"github.com/soat13/payment/internal/infra/bootstrap"
	createLinkPaymentHandler "github.com/soat13/payment/internal/infra/in/messaging/create_link"
	createPaymentHandler "github.com/soat13/payment/internal/infra/in/messaging/create_payment"
	"github.com/soat13/payment/internal/infra/in/webhook/mercado_pago"
)

type (
	App struct {
		container     *bootstrap.Container
		consumer      messaging.Consumer
		observability *observability.Components
	}
)

func NewApp(container *bootstrap.Container) *App {
	pinger := ddb.NewPinger(container.DDBClient, container.DynamodbTable)
	return &App{
		container:     container,
		consumer:      container.Consumer,
		observability: observability.Setup(container.FiberApp, pinger),
	}
}

func (a *App) Start(ctx context.Context, withHttpServer bool) {

	a.consumer.Subscribe(
		PaymentRequestQueue,
		createPaymentHandler.Handler(a.getCreatePaymentUseCase()),
	)

	a.consumer.Subscribe(
		PaymentLinkRequest,
		createLinkPaymentHandler.Handler(a.getCreateLinkUseCase()),
	)

	go a.container.Consumer.Listen(ctx)

	if withHttpServer {
		a.startFiberServer()
	}

}

func (a *App) Stop() {
	a.container.Consumer.Stop()
	observability.Shutdown(a.observability)
	_ = a.container.FiberApp.Shutdown()
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

func (a *App) getProcessPaymentStatus() *process_payment_status.ProcessPaymentStatusUseCase {
	return process_payment_status.NewProcessPaymentStatusUseCase(
		a.container.Repository,
		a.container.TopicPublisher,
		a.container.PaymentProvider,
	)
}

func (a *App) startFiberServer() {
	mercadoPagoWebhookHandler := mercado_pago.NewHandler(a.container.MercadoPagoClient, *a.getProcessPaymentStatus())

	a.container.FiberApp.Post("/webhooks/mercado-pago", mercadoPagoWebhookHandler.Handle)

	go func() {
		if err := a.container.FiberApp.Listen(":" + a.container.HttpPort); err != nil {
			log.Fatal().Err(err).
				Msg("Failed to start HTTP server")
		}
	}()
}
