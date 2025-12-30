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

	adm := cli.Admin()
	resp, err := adm.ListTopics(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range resp.Topics {
		log.Printf("topic=%s", t.Name)
	}
}
