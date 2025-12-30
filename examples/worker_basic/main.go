package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	driftq "github.com/driftq-org/DriftQ-Clients-Go/pkg/driftq"
)

func main() {
	baseURL := "http://localhost:8080"
	topic := "demo"
	group := "demo"
	owner := "worker-1"

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c, err := driftq.Dial(ctx, driftq.Config{
		BaseURL: baseURL,
	})

	if err != nil {
		log.Fatalf("Dial: %v", err)
	}

	var seen int32
	max := int32(10) // stop after N messages so the example does not run forever

	w, err := driftq.NewWorker(driftq.WorkerConfig{
		Client: c,
		Consume: driftq.ConsumeOptions{
			Topic:   topic,
			Group:   group,
			Owner:   owner,
			LeaseMS: 30_000,
		},
		Concurrency: 4,
		OnError: func(err error) {
			log.Printf("worker error: %v", err)
		},
		Handler: driftq.StepFunc(func(ctx context.Context, msg driftq.ConsumeMessage) error {
			n := atomic.AddInt32(&seen, 1)

			fmt.Printf("got #%d p=%d off=%d key=%q value=%q\n",
				n, msg.Partition, msg.Offset, msg.Key, msg.Value)

			// Simulate work stuff
			time.Sleep(50 * time.Millisecond)

			// Stop after N messages (demo)
			if n >= max {
				stop()
			}

			// return nil => Ack
			// return errors.New("fail") => Nack
			return nil
		}),
	})
	if err != nil {
		log.Fatalf("NewWorker: %v", err)
	}

	log.Printf("starting worker base_url=%s topic=%s group=%s owner=%s", baseURL, topic, group, owner)

	if err := w.Run(ctx); err != nil {
		log.Fatalf("Run: %v", err)
	}

	log.Printf("worker stopped")
}
