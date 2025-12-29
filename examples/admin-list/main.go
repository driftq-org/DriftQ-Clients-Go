package main

import (
	"context"
	"log"

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

	adm := cli.Admin()
	topics, err := adm.ListTopics(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range topics {
		log.Printf("topic=%s partitions=%d compacted=%v", t.Name, t.Partitions, t.Compacted)
	}
}
