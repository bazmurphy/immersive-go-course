package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func AdminClient() (*kadm.Client, error) {
	seeds := strings.Split(seedsFlag, ",")

	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new admin client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping the cluster: %w", err)
	}

	return kadm.NewClient(client), nil
}

func DeleteTopics(adminClient *kadm.Client, topics ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adminClient.DeleteTopics(ctx, topics...)
	if err != nil {
		return fmt.Errorf("error: failed to delete topics: %w", err)
	}

	PrintTopicsDeleted(topics)

	return nil
}

func CreateTopics(adminClient *kadm.Client, partitions int32, replicationFactor int16, configs map[string]*string, topics ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adminClient.CreateTopics(ctx, partitions, replicationFactor, configs, topics...)
	if err != nil {
		return fmt.Errorf("error: failed to create topics: %w", err)
	}

	PrintTopicsCreated(topics)

	return nil
}

func GetMetadata(adminClient *kadm.Client, topics ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata, err := adminClient.Metadata(ctx, topics...)
	if err != nil {
		return fmt.Errorf("error: failed to get topics metadata: %w", err)
	}

	PrintMetaData(metadata)

	return nil
}
