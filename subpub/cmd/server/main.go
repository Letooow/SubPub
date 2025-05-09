package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"subpub/internal/api"
	"subpub/internal/gateways"
	"syscall"
)

func main() {
	pubSub := gateways.NewPubSub(api.NewSubPub())
	var host string
	var port int
	h, ok := os.LookupEnv("HTTP_HOST")
	if !ok {
		h = "localhost"
	}
	host = h
	p, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		p = "8080"
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		port = 8080
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGTERM, syscall.SIGINT)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		s := <-sigQuit
		_, err := fmt.Printf("capturet signal: %v\n", s)
		//cancel()
		return err
	})

	server := gateways.NewSubPubServer(pubSub, gateways.WithHost(host), gateways.WithPort(uint16(port)))

	eg.Go(func() error {
		return server.Run(ctx)
	})

	if err = eg.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
