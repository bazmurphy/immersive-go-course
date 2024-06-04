package main

import (
	"log"

	"github.com/twmb/franz-go/pkg/kadm"
)

const (
	yellow  = "\x1b[33m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	reset   = "\x1b[0m"
)

func PrintTopicsDeleted(topics []string) {
	log.Printf("Deleted '%s' deleted\n", topics)
	log.Printf("Topics:\n")
}

func PrintTopicsCreated(topics []string) {
	log.Printf("topics '%s' created\n", topics)
}

func PrintMetaData(metadata kadm.Metadata) {

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
