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
	group := "group1"
	owner := "worker-1"

	msgs, errs, err := cli.ConsumeStream(ctx, driftq.ConsumeOptions{
		Topic:   topic,
		Group:   group,
		Owner:   owner,
		LeaseMS: 5000,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case m, ok := <-msgs:
			if !ok {
				log.Println("stream closed")
				return
			}

			log.Printf("[consume] partition=%d offset=%d key=%s value=%s attempts=%d",
				m.Partition, m.Offset, m.Key, m.Value, m.Attempts)

			if m.Attempts == 1 {
				log.Println("[nack] simulate failure, message will redeliver...")
				err := cli.Nack(ctx, driftq.NackRequest{
					Topic:     "demo",
					Group:     "group1",
					Owner:     "worker-1",
					Partition: m.Partition,
					Offset:    m.Offset,
					Reason:    "testing redelivery",
				})
				if err != nil {
					log.Fatalf("nack error: %v", err)
				}
				continue
			}

			log.Println("[ack] ok, finalizing message")
			err := cli.Ack(ctx, driftq.AckRequest{
				Topic:     "demo",
				Group:     "group1",
				Owner:     "worker-1",
				Partition: m.Partition,
				Offset:    m.Offset,
			})
			if err != nil {
				log.Fatalf("ack error: %v", err)
			}

		case err := <-errs:
			if err != nil {
				log.Fatalf("stream error: %v", err)
			}
		}
	}
}
