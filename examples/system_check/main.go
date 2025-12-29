package main

import (
	"context"
	"log"

	"github.com/driftq-org/DriftQ-Clients-Go/pkg/driftq"
)

func main() {
	ctx := context.Background()

	cli, err := driftq.Dial(ctx, driftq.Config{
		BaseURL: "http://localhost:8080",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	hz, err := cli.Healthz(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ver, err := cli.Version(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("healthz=%s version=%s commit=%s wal_enabled=%v",
		hz.Status, ver.Version, ver.Commit, ver.WalEnabled)
}
