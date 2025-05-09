package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func ProducerClient() (*kgo.Client, error) {
	seeds := strings.Split(seedsFlag, ",")

	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
	)
	if err != nil {
		return nil, fmt.Errorf("error: failed to create new client: %w", err)
	}
	// defer clientClusterA.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: failed to ping cluster: %w", err)
	}

	return client, nil
}
