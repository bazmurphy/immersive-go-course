// go run . -topic=cluster-a

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seedsFlag string
	topicFlag string
)

type CustomCronJobValue struct {
	ID            string `json:"id"`
	Schedule      string `json:"schedule"`
	Command       string `json:"command"`
	Topic         string `json:"topic"`
	RetryAttempts int    `json:"retry_attempts"`
	RetryDelay    int    `json:"retry_delay"`
}

func main() {
	flag.StringVar(&seedsFlag, "seeds", "kafka:19092", "the kafka broker addresses")
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

	client, err := ConsumerClient()
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

	log.Println("starting metrics server...")

	metricsServer := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		err := metricsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("error: failed to start metrics server")
		}
	}()

	// --------------------------------------------

	log.Println("consumer (polling loop) starting...")

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

			// metrics
			cronJobsConsumed.Inc()

			PrintConsumedRecord(record)

			var cronJobValue CustomCronJobValue

			err := json.Unmarshal(record.Value, &cronJobValue)
			if err != nil {
				log.Printf("error: failed to unmarshal json value: %v\n", err)
				continue
			}

			executionStartTime := time.Now()
			executionDelayStart := executionStartTime.Sub(record.Timestamp).Seconds()

			cmd := exec.Command("sh", "-c", cronJobValue.Command)

			output, err := cmd.Output()
			if err != nil {
				log.Printf("error: failed to execute cron job command: %v\n", err)

				// metrics
				cronJobsFailedExecution.Inc()

				RetryFailedCronJob(client, record, cronJobValue)
			} else {
				executionEndTime := time.Now()

				// these need to be a float64 for prometheus
				executionDelayEnd := executionEndTime.Sub(record.Timestamp).Seconds()
				executionDuration := executionEndTime.Sub(executionStartTime).Seconds()

				// metrics
				cronJobsExecuted.Inc()
				cronJobExecutionDuration.WithLabelValues(cronJobValue.Topic).Observe(executionDuration)
				cronJobExecutionDelayStart.WithLabelValues(cronJobValue.Topic).Observe(executionDelayStart)
				cronJobExecutionDelayEnd.WithLabelValues(cronJobValue.Topic).Observe(executionDelayEnd)

				PrintCommandOutput(output, executionDuration)
			}
		}
	}

	// --------------------------------------------

	err = metricsServer.Shutdown(context.Background()) // TODO: should i use ctx from above for this?
	if err != nil {
		log.Printf("error: failed to shut down metrics server: %v\n", err)
	}

	log.Println("metrics server shut down")
}
