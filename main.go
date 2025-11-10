package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config := ProcessInput()
	err := NewExtractor(config).Process(ctx)
	if err != nil {
		os.Exit(1)
	}
}
