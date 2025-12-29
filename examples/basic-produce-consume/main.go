package main

import (
	"context"
	"log"
	"time"

	// TODO go get the below
	"github.com/driftq-org/DriftQ-Clients-Go/pkg/driftq"
)

func main() {
	ctx := context.Background()
	cli, err := driftq.Dial(ctx, driftq.Config{BaseURL: "http://localhost:8080"})

	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	stop, err := cli.Consumer("orders", "orders-workers").Start(ctx, func(m *driftq.Message) error {
		log.Printf("consumed (stub): %s", string(m.Value))
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = stop() }()

	if err := cli.Producer("orders").Send(ctx, &driftq.Message{
		Key: "k", Value: []byte("hello"),
	}); err != nil {
		log.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
}
