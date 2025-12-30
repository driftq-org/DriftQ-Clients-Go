package main

import (
	"context"
	"log"
	"time"

	"github.com/driftq-org/DriftQ-Clients-Go/pkg/driftq"
)

func main() {
	ctx := context.Background()
	cli, err := driftq.Dial(ctx, driftq.Config{BaseURL: "http://localhost:8080"})

	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	const (
		topic = "demo"
		group = "group1"
		owner = "worker-1"
	)

	msgs, errs, err := cli.ConsumeStream(ctx, driftq.ConsumeOptions{
		Topic:   topic,
		Group:   group,
		Owner:   owner,
		LeaseMS: 5000,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Nack the first time we see a message, then wait for redelivery and Ack
	nackedOnce := false

	for {
		select {
		case m, ok := <-msgs:
			if !ok {
				log.Println("stream closed")
				return
			}

			log.Printf("[consume] partition=%d offset=%d key=%q value=%q attempts=%d",
				m.Partition, m.Offset, m.Key, m.Value, m.Attempts)

			if !nackedOnce {
				nackedOnce = true

				log.Println("[nack] simulate failure, message should redeliver")
				if err := cli.Nack(ctx, driftq.NackRequest{
					Topic:     topic,
					Group:     group,
					Owner:     owner,
					Partition: m.Partition,
					Offset:    m.Offset,
					Reason:    "testing redelivery",
				}); err != nil {
					log.Fatalf("nack error: %v", err)
				}

				// small pause just so logs are easier to read
				time.Sleep(500 * time.Millisecond)
				continue
			}

			log.Println("[test] sleeping past lease...")
			time.Sleep(2 * time.Second)

			log.Println("[ack] ok, finalizing message")
			if err := cli.Ack(ctx, driftq.AckRequest{
				Topic:     topic,
				Group:     group,
				Owner:     owner,
				Partition: m.Partition,
				Offset:    m.Offset,
			}); err != nil {
				log.Fatalf("ack error: %v", err)
			}

			log.Println("done")
			return

		case err := <-errs:
			if err != nil {
				log.Fatalf("stream error: %v", err)
			}
		}
	}
}
