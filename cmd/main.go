package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/soat13/oficina-utils/pkg/observability"
	"github.com/soat13/payment/internal/infra"
	"github.com/soat13/payment/internal/infra/bootstrap"
)

func main() {
	ctx := context.Background()
	container, err := bootstrap.NewContainer(ctx, nil)
	if err != nil {
		panic(err)
	}

	obs := observability.Setup(nil, nil)
	defer observability.Shutdown(obs)
	container.Metrics = obs.Metrics

	application := infra.NewApp(container)

	application.Start(ctx, true)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverShutdownWaitGroup := sync.WaitGroup{}
	serverShutdownWaitGroup.Go(func() {
		<-quit
		application.Stop()
		ctx.Done()
	})

	serverShutdownWaitGroup.Wait()
}
