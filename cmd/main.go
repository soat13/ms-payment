package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	app "github.com/soat13/payment/internal"
	"github.com/soat13/payment/internal/infra/bootstrap"
)

func main() {
	ctx := context.Background()
	container := bootstrap.NewContainer(bootstrap.DefaultEnv)
	application, err := app.New(ctx, container)

	if err != nil {
		panic(err)
	}

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
