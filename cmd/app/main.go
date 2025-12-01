package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/guarref/link-checking-service/internal/app"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	application, err := app.New()
	if err != nil {
		log.Fatalf("error creating new application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
