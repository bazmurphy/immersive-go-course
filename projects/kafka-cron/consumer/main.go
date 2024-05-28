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
	ID            string `json:"id"`
	Schedule      string `json:"schedule"`
	Command       string `json:"command"`
	Cluster       string `json:"cluster"`
	RetryAttempts int    `json:"retry_attempts"`
}

func main() {
	flag.StringVar(&seedsFlag, "seeds", "", "the kafka broker addresses")
	flag.StringVar(&topicFlag, "topic", "", "the name of the topic")
	flag.BoolVar(&retryTopicFlag, "retry-topic", false, "whether to consume from the retry topic")

	flag.Parse()
	// log.Println("DEBUG | seedsFlag:", seedsFlag, "topicFlag:", topicFlag, "retryTopicFlag:", retryTopicFlag)

	// TODO: handle the flag errors more specifically
	if seedsFlag == "" || topicFlag == "" {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	log.Println("connecting to the cluster...")

	client, err := clusterConnection()
	if err != nil {
		log.Fatalf("error: failed to establish connection to the cluster: %v", err)
	}
	defer client.Close()

	log.Println("connection established to the cluster...")

	// --------------------------------------------

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
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

			printConsumedRecord(record)

			var cronJobValue CustomCronJobValue

			err := json.Unmarshal(record.Value, &cronJobValue)
			if err != nil {
				log.Printf("error: failed to unmarshal json value: %v\n", err)
				continue
			}

			cmd := exec.Command("sh", "-c", cronJobValue.Command)

			output, err := cmd.Output()
			if err != nil {
				log.Printf("error: failed to execute cron job command: %v\n", err)

				// (!) Part 3 Retry Logic:
				if cronJobValue.RetryAttempts == 0 {
					log.Printf("error: out of retry attempts: %v\n", err)
					continue
				}

				if cronJobValue.RetryAttempts > 0 {
					log.Printf("attempting retry %d...\n", cronJobValue.RetryAttempts)

					retryCronJobValue := cronJobValue

					// 1. reduce the number of attempts
					retryCronJobValue.RetryAttempts--

					// 2. convert the retry value into json
					retryValueJSON, err := json.Marshal(retryCronJobValue)
					if err != nil {
						log.Printf("error: failed to marshal retry cron job value to json: %v\n", err)
						continue
					}

					// 3. define the retry topic
					retryTopic := topicFlag + "-retry"

					// 4. create the new retry record
					retryRecord := &kgo.Record{
						Topic: retryTopic,
						Key:   record.Key,
						Value: retryValueJSON,
					}

					log.Println("producing new retry record...")

					// TODO: fix the context here
					client.Produce(context.Background(), retryRecord, func(_ *kgo.Record, err error) {
						if err != nil {
							log.Printf("error: failed to produce retry record: %v\n", err)
							return
						}
						log.Printf("produced retry record:\n\tcluster:%s\n\ttopic:%s\n\tpartition:%d\n\toffset:%d\n\ttimestamp:%v\n\tkey:%s\n\tvalue:%s\n", cronJobValue.Cluster, retryRecord.Topic, retryRecord.Partition, retryRecord.Offset, retryRecord.Timestamp, retryRecord.Key, retryRecord.Value)
					})
				}
			} else {
				printCommandOutput(output)
			}
		}
	}
}

func clusterConnection() (*kgo.Client, error) {
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
		return nil, fmt.Errorf("error: failed to create new client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: failed to ping the cluster: %w", err)
	}

	return client, nil
}

func printConsumedRecord(cj *kgo.Record) {
	red := "\x1b[31m"
	green := "\x1b[32m"
	yellow := "\x1b[33m"
	blue := "\x1b[34m"
	magenta := "\x1b[35m"
	reset := "\x1b[0m"

	log.Printf(
		"%sConsumed Record%s:\n\t%sTopic%s: %s%s%s\n\t%sPartition%s: %s%d%s\n\t%sOffset%s: %s%d%s\n\t%sTimestamp%s: %s%v%s\n\t%sKey%s: %s%s%s\n\t%sValue%s: %s%s%s\n",
		yellow, reset,
		blue, reset, red, cj.Topic, reset,
		blue, reset, magenta, cj.Partition, reset,
		blue, reset, magenta, cj.Offset, reset,
		blue, reset, magenta, cj.Timestamp, reset,
		blue, reset, green, string(cj.Key), reset,
		blue, reset, green, string(cj.Value), reset,
	)
}

func printCommandOutput(output []byte) {
	yellow := "\x1b[33m"
	magenta := "\x1b[35m"
	reset := "\x1b[0m"

	log.Printf("%sCommand Output%s:\n\t%s%s%s\n",
		yellow, reset,
		magenta, output, reset)
}
