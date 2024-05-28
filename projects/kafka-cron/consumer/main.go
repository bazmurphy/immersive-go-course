// go run . -topic=cluster-a

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
	Topic         string `json:"topic"`
	RetryAttempts int    `json:"retry_attempts"`
}

func main() {
	flag.StringVar(&seedsFlag, "seeds", "localhost:9092", "the kafka broker addresses")
	flag.StringVar(&topicFlag, "topic", "", "the name of the topic")

	flag.Parse()
	// log.Println("DEBUG | seedsFlag:", seedsFlag, "topicFlag:", topicFlag)

	// TODO: handle the flag errors more specifically
	if seedsFlag == "" || topicFlag == "" {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	log.Println("connecting to the cluster...")

	client, err := kafkaClient()
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

				if cronJobValue.RetryAttempts > 0 {
					// 1. make a copy of the original cronjob value
					retryCronJobValue := cronJobValue

					retryTopic := retryCronJobValue.Topic
					// 2. if it's the first time we are retrying adjust the topic
					if retryCronJobValue.RetryAttempts == 3 {
						retryTopic = cronJobValue.Topic + "-retry"
					}

					// 3. update the cron job value topic
					retryCronJobValue.Topic = retryTopic

					// 4. reduce the number of retry attempts
					retryCronJobValue.RetryAttempts--

					// 5. convert the retry value into json
					retryValueJSON, err := json.Marshal(retryCronJobValue)
					if err != nil {
						log.Printf("error: failed to marshal retry cron job value to json: %v\n", err)
						continue
					}

					// 6. create the new retry record
					retryRecord := &kgo.Record{
						Topic: retryTopic,
						Key:   record.Key,
						Value: retryValueJSON,
					}

					// attempt 3 = 5 seconds, attempt 2 = 10 seconds, attempt 1 = 15 seconds
					baseSeconds := 20
					sleepSeconds := baseSeconds - cronJobValue.RetryAttempts*5
					log.Printf("sleeping for %d seconds before retrying...\n", sleepSeconds)

					time.Sleep(time.Duration(sleepSeconds) * time.Second)

					// 8. produce the retry record to the -retry topic

					// TODO: fix the context here
					client.Produce(context.Background(), retryRecord, func(_ *kgo.Record, err error) {
						if err != nil {
							log.Printf("error: failed to produce retry record: %v\n", err)
							return
						}

						printRetryProducedRecord(retryRecord)
					})
				}

				if cronJobValue.RetryAttempts == 0 {
					log.Printf("error: out of retry attempts: %v\n", err)
					continue
				}
			} else {
				printCommandOutput(output)
			}
		}
	}
}

func kafkaClient() (*kgo.Client, error) {
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

func printRetryProducedRecord(cj *kgo.Record) {
	red := "\x1b[31m"
	green := "\x1b[32m"
	// yellow := "\x1b[33m"
	blue := "\x1b[34m"
	magenta := "\x1b[35m"
	reset := "\x1b[0m"

	log.Printf(
		"%sProduced Retry Record%s:\n\t%sTopic%s: %s%s%s\n\t%sPartition%s: %s%d%s\n\t%sOffset%s: %s%d%s\n\t%sTimestamp%s: %s%v%s\n\t%sKey%s: %s%s%s\n\t%sValue%s: %s%s%s\n",
		red, reset,
		blue, reset, red, cj.Topic, reset,
		blue, reset, magenta, cj.Partition, reset,
		blue, reset, magenta, cj.Offset, reset,
		blue, reset, magenta, cj.Timestamp, reset,
		blue, reset, green, string(cj.Key), reset,
		blue, reset, green, string(cj.Value), reset,
	)
}
