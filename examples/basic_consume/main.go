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

			// ACK using the fields Core requires: topic+group+owner+partition+offset
			if err := cli.Ack(ctx, driftq.AckRequest{
				Topic:     topic,
				Group:     group,
				Owner:     owner,
				Partition: m.Partition,
				Offset:    m.Offset,
			}); err != nil {
				log.Printf("[ack] failed partition=%d offset=%d: %v", m.Partition, m.Offset, err)
			} else {
				log.Printf("[ack] ok partition=%d offset=%d", m.Partition, m.Offset)
			}

		case err := <-errs:
			if err != nil {
				log.Fatalf("stream error: %v", err)
			}
		}
	}
}
