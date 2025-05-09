package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func RetryFailedCronJob(client *kgo.Client, record *kgo.Record, cronJobValue CustomCronJobValue) {
	// 1. if it's the first time we are retrying, then produce it to the retry topic (eg. cluster-a-retry)
	// [this conditional check is hacky and wrong - please suggest better]
	if !strings.Contains(record.Topic, "retry") {
		// define the retry topic for where to produce the retry record to
		retryTopic := record.Topic + "-retry"

		// copy the original value and adjust the topic
		retryCronJobValue := cronJobValue
		retryCronJobValue.Topic = retryTopic

		retryValueJSON, err := json.Marshal(retryCronJobValue)
		if err != nil {
			log.Printf("error: failed to marshal retry cron job value to json: %v\n", err)
			return
		}

		retryRecord := &kgo.Record{
			Topic: retryTopic,
			Key:   record.Key,
			Value: retryValueJSON,
		}

		log.Printf("warn: producing cron job to retry topic: %s\n", retryRecord.Topic)

		client.Produce(context.Background(), retryRecord, func(_ *kgo.Record, err error) {
			if err != nil {
				log.Printf("error: failed to produce retry record: %v\n", err)
				return
			}

			PrintRetryProducedRecord(retryRecord)
		})

		return
	}

	// 2. if we have no more retry attempts
	if cronJobValue.RetryAttempts == 0 {
		// metrics
		cronJobsRetryFailed.Inc()

		log.Printf("error: out of retry attempts: cronjob failed to run: %v\n", cronJobValue)
		return
	}

	// 3. if we are in the retry "loop"
	// copy the original value and decrement the retry attempts
	retryCronJobValue := cronJobValue
	retryCronJobValue.RetryAttempts--

	retryValueJSON, err := json.Marshal(retryCronJobValue)
	if err != nil {
		log.Printf("error: failed to marshal retry cron job value to json: %v\n", err)
		return
	}

	retryRecord := &kgo.Record{
		Topic: record.Topic,
		Key:   record.Key,
		Value: retryValueJSON,
	}

	retryDelay := time.Duration(retryCronJobValue.RetryDelay) * time.Second

	log.Printf("warn: sleeping for %v then producing retry cron job to retry topic...\n", retryDelay)

	time.Sleep(retryDelay)

	client.Produce(context.Background(), retryRecord, func(_ *kgo.Record, err error) {
		if err != nil {
			log.Printf("error: failed to produce retry record: %v\n", err)
			return
		}

		// metrics
		cronJobsRetried.Inc()

		PrintRetryProducedRecord(retryRecord)
	})
}
