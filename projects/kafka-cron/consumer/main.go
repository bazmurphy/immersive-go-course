// go run . -seeds=localhost:9092,localhost:9093,localhost:9094 -topic=cron-topic -retry-topic=false

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seedsFlag      string
	topicFlag      string
	retryTopicFlag bool
)

type CustomCronJobValue struct {
	Cluster       string `json:"cluster"`
	ID            string `json:"id"`
	Schedule      string `json:"schedule"`
	Command       string `json:"command"`
	RetryAttempts int    `json:"retry_attempts"`
}

func main() {
	flag.StringVar(&seedsFlag, "seeds", "", "the kafka broker addresses")
	flag.StringVar(&topicFlag, "topic", "", "the name of the topic")
	flag.BoolVar(&retryTopicFlag, "retry-topic", false, "whether to consume from the retry topic")

	flag.Parse()
	log.Println("DEBUG | seedsFlag:", seedsFlag, "topicFlag:", topicFlag, "retryTopicFlag:", retryTopicFlag)

	if seedsFlag == "" || topicFlag == "" {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	seeds := strings.Split(seedsFlag, ",")

	topic := topicFlag
	if retryTopicFlag {
		topic = topicFlag + "-retry"
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(topic+"-consumer-group"),
		kgo.ConsumeTopics(topic),
		// kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	if err != nil {
		log.Fatalf("error: failed to create new client: %v\n", err)
	}
	defer client.Close()

	// --------------------------------------------

	log.Println("connecting to the cluster...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx)
	if err != nil {
		log.Fatalf("error: failed to ping the cluster: %v\n", err)
	}

	log.Println("connection established with cluster...")

	// --------------------------------------------

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-signalChannel
		cancel() // this cancels the context for the polling loop below
	}()

	// --------------------------------------------

	log.Println("polling loop starting...")

	// polling loop (outer)
	for {
		fetches := client.PollFetches(ctx) // use the context from above to allow cancellation of the polling loop

		if errs := fetches.Errors(); len(errs) > 0 {
			if errors.Is(fetches.Err(), context.Canceled) {
				fmt.Println("context cancelled, shutting down...")
				break
			} else if errors.Is(fetches.Err(), kgo.ErrClientClosed) {
				fmt.Println("client closed, shutting down...")
				break
			}
			fmt.Printf("error: fetches errors: %v\n", fetches.Err())
			continue
		}

		iterator := fetches.RecordIter()

		// iterator loop (inner)
		for !iterator.Done() {
			record := iterator.Next()

			log.Printf("consumed record:\n\ttopic:%s\n\tpartition:%d\n\toffset:%d\n\ttimestamp:%v\tkey:%s\n\tvalue:%s\n", record.Topic, record.Partition, record.Offset, record.Timestamp, record.Key, record.Value)

			var cronJobValue CustomCronJobValue

			err := json.Unmarshal(record.Value, &cronJobValue)
			if err != nil {
				log.Printf("error: failed to unmarshal json value: %v\n", err)
				continue
			}

			cmd := exec.Command("sh", "-c", cronJobValue.Command)

			output, err := cmd.Output()
			if err != nil {
				log.Printf("error: failed to execute command: %v\n", err)
				// TODO: add Part 3 retry logic here...

			} else {
				log.Printf("success: command output: %s\n", output)
			}
		}
	}
}
