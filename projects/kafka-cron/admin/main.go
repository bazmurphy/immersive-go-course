package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seedsFlag             string
	topicFlag             string
	partitionsFlag        int
	replicationFactorFlag int
)

func main() {
	flag.StringVar(&seedsFlag, "seeds", "", "the kafka broker addresses")
	flag.StringVar(&topicFlag, "topic", "", "the name of the topic")
	flag.IntVar(&partitionsFlag, "partitions", 0, "the number of partitions")
	flag.IntVar(&replicationFactorFlag, "replication-factor", 0, "the replication factor")

	flag.Parse()

	// log.Println("DEBUG | seedsFlag:", seedsFlag, "topicFlag:", topicFlag, "partitionsFlag:", partitionsFlag, "replicationFactorFlag:", replicationFactorFlag)

	if seedsFlag == "" || topicFlag == "" || partitionsFlag == 0 || replicationFactorFlag == 0 {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	seeds := strings.Split(seedsFlag, ",")

	log.Printf("new kafka admin client starting...")

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topic := topicFlag
	partitions := int32(partitionsFlag)
	replicationFactor := int16(replicationFactorFlag)
	configs := make(map[string]*string)

	deleteTopicResponse, err := adminClient.DeleteTopic(ctx, topic)
	if err != nil {
		log.Printf("error: failed to delete topic %s: %v\n", topic, err)
	} else {
		log.Printf("success: topic '%s' deleted: %v\n", topic, deleteTopicResponse)
	}

	time.Sleep(1 * time.Second)

	createTopicResponse, err := adminClient.CreateTopic(ctx, partitions, replicationFactor, configs, topic)
	if err != nil {
		log.Printf("error: failed to create topic %s: %v\n", topic, err)
	} else {
		log.Printf("success: topic '%s' created: %v\n", topic, createTopicResponse)
	}

	time.Sleep(1 * time.Second)

	metadata, err := adminClient.Metadata(ctx, topic)
	if err != nil {
		log.Printf("error: failed to get metadata for topic %s: %v\n", topic, err)
	} else {
		log.Print("success: Metadata:\n")
		log.Printf("  Cluster: %v\n", metadata.Cluster)
		log.Printf("  Controller Broker: %v\n", metadata.Controller)
		log.Print("  Brokers:\n")
		for _, broker := range metadata.Brokers {
			log.Printf("    - ID: %d, Host: %s, Port: %d\n",
				broker.NodeID, broker.Host, broker.Port)
		}
		log.Print("  Topics:\n")
		for _, topic := range metadata.Topics.Sorted() {
			log.Printf("    - Topic: %s\n", topic.Topic)
			log.Printf("      ID: %s\n", topic.ID)
			log.Printf("      Partitions:\n")
			for _, partition := range topic.Partitions.Sorted() {
				log.Printf("        - Partition: %d\n", partition.Partition)
				log.Printf("          Leader: %d\n", partition.Leader)
				log.Printf("          Replicas: %v\n", partition.Replicas)
				log.Printf("          In Sync Replicas: %v\n", partition.ISR)
			}
		}
	}
}
