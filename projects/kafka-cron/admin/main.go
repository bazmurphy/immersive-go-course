// go run . -topics=cluster-a,cluster-b

package main

import (
	"flag"
	"log"
	"strings"
	"time"
)

var (
	seedsFlag             string
	topicsFlag            string
	partitionsFlag        int
	replicationFactorFlag int
)

func main() {
	flag.StringVar(&seedsFlag, "seeds", "kafka:19092", "the kafka broker addresses")
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

	adminClient, err := AdminClient()
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

	err = DeleteTopics(adminClient, topics...)
	if err != nil {
		log.Fatalf("error: failed to delete topics: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// --------------------------------------------

	err = CreateTopics(adminClient, partitions, replicationFactor, configs, topics...)
	if err != nil {
		log.Fatalf("error: failed to create topics: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// --------------------------------------------

	err = GetMetadata(adminClient, topics...)
	if err != nil {
		log.Fatalf("error: failed to get the metadata of the topics: %v", err)
	}
}
