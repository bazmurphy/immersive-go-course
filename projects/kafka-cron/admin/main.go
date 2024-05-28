// go run . -cluster-a-seeds=localhost:9092,localhost:9093,localhost:9094 -cluster-b-seeds=localhost:9095,localhost:9096,localhost:9097 -topic=cron-topic -partitions=3 -replication-factor=3

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	clusterASeedsFlag     string
	clusterBSeedsFlag     string
	topicFlag             string
	partitionsFlag        int
	replicationFactorFlag int
)

func main() {
	flag.StringVar(&clusterASeedsFlag, "cluster-a-seeds", "", "the kafka broker addresses of cluster a")
	flag.StringVar(&clusterBSeedsFlag, "cluster-b-seeds", "", "the kafka broker addresses of cluster b")
	flag.StringVar(&topicFlag, "topic", "", "the name of the topic")
	flag.IntVar(&partitionsFlag, "partitions", 0, "the number of partitions")
	flag.IntVar(&replicationFactorFlag, "replication-factor", 0, "the replication factor")

	flag.Parse()
	// log.Println("DEBUG | clusterASeedsFlag:", clusterASeedsFlag, "clusterBSeedsFlag:", clusterBSeedsFlag, "topicFlag:", topicFlag, "partitionsFlag:", partitionsFlag, "replicationFactorFlag:", replicationFactorFlag)

	// TODO: handle the flag errors more specifically
	if clusterASeedsFlag == "" || clusterBSeedsFlag == "" || topicFlag == "" || partitionsFlag == 0 || replicationFactorFlag == 0 {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	clusterASeeds := strings.Split(clusterASeedsFlag, ",")
	clusterBSeeds := strings.Split(clusterBSeedsFlag, ",")

	adminClientClusterA, err := createAdminClient(clusterASeeds)
	if err != nil {
		log.Fatalf("error: failed to create admin client for cluster-a: %v", err)
	}
	defer adminClientClusterA.Close()

	adminClientClusterB, err := createAdminClient(clusterBSeeds)
	if err != nil {
		log.Fatalf("error: failed to create admin client for cluster-b: %v", err)
	}
	defer adminClientClusterB.Close()

	// --------------------------------------------

	topic := topicFlag
	partitions := int32(partitionsFlag)
	replicationFactor := int16(replicationFactorFlag)
	configs := make(map[string]*string)

	retryTopic := topic + "-retry"

	err = deleteTopics(adminClientClusterA, topic, retryTopic)
	if err != nil {
		log.Fatalf("error: failed to delete the topics from cluster-a: %v", err)
	}
	err = deleteTopics(adminClientClusterB, topic, retryTopic)
	if err != nil {
		log.Fatalf("error: failed to delete the topics from cluster-b: %v", err)
	}

	time.Sleep(1 * time.Second)

	// --------------------------------------------

	err = createTopics(adminClientClusterA, partitions, replicationFactor, configs, topic, retryTopic)
	if err != nil {
		log.Fatalf("error: failed to create the topics on cluster-a: %v", err)
	}
	err = createTopics(adminClientClusterB, partitions, replicationFactor, configs, topic, retryTopic)
	if err != nil {
		log.Fatalf("error: failed to create the topics on cluster-b: %v", err)
	}

	time.Sleep(1 * time.Second)

	// --------------------------------------------

	err = getMetadata(adminClientClusterA, topic, retryTopic)
	if err != nil {
		log.Fatalf("error: failed to get the metadata of cluster-a: %v", err)
	}
	err = getMetadata(adminClientClusterB, topic, retryTopic)
	if err != nil {
		log.Fatalf("error: failed to get the metadata of cluster-b: %v", err)
	}

	// --------------------------------------------
}

func createAdminClient(seeds []string) (*kadm.Client, error) {
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

func deleteTopics(adminClient *kadm.Client, topics ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adminClient.DeleteTopics(ctx, topics...)
	if err != nil {
		return fmt.Errorf("error: failed to delete topics: %w", err)
	}

	log.Printf("topics '%s' deleted\n", topics)
	return nil
}

func createTopics(adminClient *kadm.Client, partitions int32, replicationFactor int16, configs map[string]*string, topics ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adminClient.CreateTopics(ctx, partitions, replicationFactor, configs, topics...)
	if err != nil {
		return fmt.Errorf("error: failed to create topics: %w", err)
	}

	log.Printf("topics '%s' created\n", topics)
	return nil
}

func getMetadata(adminClient *kadm.Client, topics ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata, err := adminClient.Metadata(ctx, topics...)
	if err != nil {
		return fmt.Errorf("error: failed to get metadata: %w", err)
	}

	PrintMetaData(metadata)
	return nil
}

func PrintMetaData(metadata kadm.Metadata) {
	log.Print("Metadata:\n")
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
