package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	log.Printf("new kafka client starting...")

	seeds := []string{"localhost:9092"}
	// TODO: pass seeds in as a flag/environment variable?

	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("cron-consumer-group"),
		kgo.ConsumeTopics("cron-topic"),
		// kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	if err != nil {
		log.Fatalf("error: failed to create new kafka client: %v\n", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// TODO: but this context is never cancelled...

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChannel
		log.Printf("received signal, closing client...\n")
		// cancel() // TODO: should we cancel the context here? it causes issues in the for !iterator.Done() loop
		client.Close()
	}()

	for {
		fetches := client.PollFetches(ctx)

		if fetches.IsClientClosed() {
			log.Printf("client closed, shutting down...\n")
			return
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			log.Printf("error: client fetches errors: %v", errs)
			// TODO: how to handle the errors?
		}

		iterator := fetches.RecordIter()

		for !iterator.Done() {
			record := iterator.Next()
			log.Printf("consumed record:\n\tkey:%s\n\tvalue:%s\n\ttopic:%s\n\tpartition:%d\n\toffset:%d\n\ttimestamp:%v", record.Key, record.Value, record.Topic, record.Partition, record.Offset, record.Timestamp)

			var cronJobValue CronJobValue

			err := json.Unmarshal(record.Value, &cronJobValue)
			if err != nil {
				log.Printf("error: failed to unmarshal json value: %v\n", err)
				continue
			}

			cmd := exec.Command("sh", "-c", cronJobValue.Command)

			output, err := cmd.Output()
			if err != nil {
				log.Printf("error: failed to execute command: %v\n", err)
			} else {
				log.Printf("success: command output: %s\n", output)
			}

			// TODO: this is related to how cancelling the context affects this loop...

			// if ctx.Err() != nil {
			// 	break
			// }
		}
	}
}

type CronJobValue struct {
	ID       string `json:"id"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
}
