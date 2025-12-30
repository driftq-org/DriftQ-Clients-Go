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

	_, _ = cli.Admin().CreateTopic(ctx, topic, 1)

	pr, err := cli.Produce(ctx, driftq.ProduceRequest{
		Topic: topic,
		Value: "hello from DriftQ-Clients-Go",
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("produced status=%s topic=%s", pr.Status, pr.Topic)
	log.Printf("next: run a consumer stream and Ack/Nack using partition+offset+owner.")
}
