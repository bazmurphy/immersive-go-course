package main

import (
	"context"
	"log"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	seeds := []string{"localhost:9092"}
	// TODO: pass seeds in as a flag/environment variable?

	var adminClient *kadm.Client
	{
		client, err := kgo.NewClient(
			kgo.SeedBrokers(seeds...),
		)
		if err != nil {
			log.Fatalf("error: failed to create new admin client: %v\n", err)
		}
		defer client.Close()

		adminClient = kadm.NewClient(client)
	}

	ctx := context.Background()

	topic := "cron-topic"
	partitions := int32(3)
	replicationFactor := int16(1)
	// TODO: pass topic/partitions/replicationFactor in as flags/environment variables?
	configs := make(map[string]*string)

	// deleteTopicResponse, err := adminClient.DeleteTopic(ctx, topic)
	// if err != nil {
	// 	log.Fatalf("error: failed to delete topic %s: %v\n", topic, deleteTopicResponse.Err)
	// }
	// log.Printf("success: topic '%s' deleted\n", topic)

	// NOTE: this is a very manual way to create the topic
	createTopicResponse, err := adminClient.CreateTopic(ctx, partitions, replicationFactor, configs, topic)
	if err != nil {
		log.Fatalf("error: failed to create topic %s: %v\n", topic, createTopicResponse.Err)
	}
	log.Printf("success: topic '%s' created\n", topic)
}
