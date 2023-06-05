package main

import (
	"context"
	"github.com/itzloop/iot-vkube/internal/agent"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		s := <-sig
		logrus.WithField("signal", s.String()).Info("received interrupt, quitting gracefully")
		cancel()

		s = <-sig
		logrus.WithField("signal", s.String()).Info("force quit")
		os.Exit(0)
	}()

	agent.NewService(store.NewLocalStoreImpl(), ":8080", nil, []string{"localhost:5000"}).Start(ctx)
}
