package main

import (
	"context"
	"log"

	"github.com/driftq-org/DriftQ-Clients-Go/pkg/driftq"
)

func main() {
	ctx := context.Background()

	cli, err := driftq.Dial(ctx, driftq.Config{BaseURL: "http://localhost:8080"})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	topic := "demo"

	// Create topic (ignore error if it already exists)
	_ = cli.Admin().CreateTopic(ctx, topic)

	pr, err := cli.Produce(ctx, driftq.ProduceRequest{
		Topic: topic,
		Value: "hello from DriftQ-Clients-Go",
	})

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("produced status=%s topic=%s", pr.Status, pr.Topic)

	// Ack/Nack require owner + id from a consume lease.
	log.Printf("next: implement ConsumeStream; then you can Ack/Nack this id with owner.")
}
