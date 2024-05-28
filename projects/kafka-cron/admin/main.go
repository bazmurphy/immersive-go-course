// go run . -topics=cluster-a,cluster-b

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
	seedsFlag             string
	topicsFlag            string
	partitionsFlag        int
	replicationFactorFlag int
)

func main() {
	flag.StringVar(&seedsFlag, "seeds", "localhost:9092", "the kafka broker addresses")
	flag.StringVar(&topicsFlag, "topics", "", "the name of the topics")
	flag.IntVar(&partitionsFlag, "partitions", 1, "the number of partitions")
	flag.IntVar(&replicationFactorFlag, "replication-factor", 1, "the replication factor")

	flag.Parse()
	// log.Println("DEBUG | seedsFlag:", seedsFlag, "topicsFlag:", topicsFlag, "partitionsFlag:", partitionsFlag, "replicationFactorFlag:", replicationFactorFlag)

	// TODO: handle the flag errors more specifically
	if seedsFlag == "" || topicsFlag == "" || partitionsFlag == 0 || replicationFactorFlag == 0 {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	adminClient, err := kafkaAdminClient()
	if err != nil {
		log.Fatalf("error: failed to establish admin client connection: %v", err)
	}
	defer adminClient.Close()

	// --------------------------------------------

	topics := strings.Split(topicsFlag, ",")
	for _, topic := range topics {
		topics = append(topics, topic+"-retry")
	}
	partitions := int32(partitionsFlag)
	replicationFactor := int16(replicationFactorFlag)
	configs := make(map[string]*string)

	// --------------------------------------------

	err = deleteTopics(adminClient, topics...)
	if err != nil {
		log.Fatalf("error: failed to delete topics: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// --------------------------------------------

	err = createTopics(adminClient, partitions, replicationFactor, configs, topics...)
	if err != nil {
		log.Fatalf("error: failed to create topics: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// --------------------------------------------

	err = getMetadata(adminClient, topics...)
	if err != nil {
		log.Fatalf("error: failed to get the metadata of the topics: %v", err)
	}
}

func kafkaAdminClient() (*kadm.Client, error) {
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
		return fmt.Errorf("error: failed to get topics metadata: %w", err)
	}

	printMetaData(metadata)

	return nil
}

func printMetaData(metadata kadm.Metadata) {
	yellow := "\x1b[33m"
	magenta := "\x1b[35m"
	cyan := "\x1b[36m"
	reset := "\x1b[0m"

	log.Printf("%sMetadata:%s\n", yellow, reset)
	log.Printf("  %sCluster%s: %s%v%s\n", cyan, reset, magenta, metadata.Cluster, reset)
	log.Printf("  %sController Broker%s: %s%v%s\n", cyan, reset, magenta, metadata.Controller, reset)
	log.Printf("  %sBrokers:%s\n", cyan, reset)
	for _, broker := range metadata.Brokers {
		log.Printf("    - %sID%s: %s%d%s, %sHost%s: %s%s%s, %sPort%s: %s%d%s\n",
			cyan, reset, magenta, broker.NodeID, reset,
			cyan, reset, magenta, broker.Host, reset,
			cyan, reset, magenta, broker.Port, reset)
	}
	log.Printf("  %sTopics:%s\n", cyan, reset)
	for _, topic := range metadata.Topics.Sorted() {
		log.Printf("    - %sTopic%s: %s%s%s\n", cyan, reset, magenta, topic.Topic, reset)
		log.Printf("      %sID%s: %s%s%s\n", cyan, reset, magenta, topic.ID, reset)
		log.Printf("      %sPartitions:%s\n", cyan, reset)
		for _, partition := range topic.Partitions.Sorted() {
			log.Printf("        - %sPartition%s: %s%d%s\n", cyan, reset, magenta, partition.Partition, reset)
			log.Printf("          %sLeader%s: %s%d%s\n", cyan, reset, magenta, partition.Leader, reset)
			log.Printf("          %sReplicas%s: %s%v%s\n", cyan, reset, magenta, partition.Replicas, reset)
		}
	}
}
