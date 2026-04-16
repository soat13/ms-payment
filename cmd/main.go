package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/soat13/payment/internal/infra/bootstrap"
)

func main() {
	ctx := context.Background()
	container, err := bootstrap.NewContainer(ctx, bootstrap.DefaultEnvs)
	if err != nil {
		panic(err)
	}

	application := bootstrap.NewApp(
		container.Repository,
		container.TopicPublisher,
		container.QueueSender,
		container.Consumer,
	)

	application.Start(ctx)

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
