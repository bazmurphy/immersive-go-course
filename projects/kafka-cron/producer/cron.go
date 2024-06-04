package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/twmb/franz-go/pkg/kgo"
)

type CustomCronJob struct {
	ID            string
	Schedule      string
	Command       string
	Cluster       string
	RetryAttempts int
	RetryDelay    int
	Topic         string
	Client        *kgo.Client
}

type CustomCronJobValue struct {
	ID            string `json:"id"`
	Schedule      string `json:"schedule"`
	Command       string `json:"command"`
	Topic         string `json:"topic"`
	RetryAttempts int    `json:"retry_attempts"`
	RetryDelay    int    `json:"retry_delay"`
}

func ParseCronTabFile(cronTabFile *os.File) ([]CustomCronJob, error) {
	var customCronJobs []CustomCronJob

	scanner := bufio.NewScanner(cronTabFile)

	var lineCount int

	for scanner.Scan() {
		line := scanner.Text()

		lineCount++

		// TODO: (as always) there are infinite possibilities of what could be wrong with each line...

		// skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 9 {
			log.Printf("warn: crontab file line %d invalid: expect a minimum of 9 fields)\n", lineCount)
			continue
		}
		// TODO: there is no guarantee just because we have 9+ fields that it is correct

		schedule := strings.Join(fields[:6], " ")

		command := strings.Join(fields[6:len(fields)-2], " ")

		topic := fields[len(fields)-3] // TODO: this could panic and/or be totally incorrect

		if topic != "cluster-a" && topic != "cluster-b" {
			log.Printf("warn: crontab file line %d invalid: topic name\n", lineCount)
			continue
		}

		retryAttemptsString := fields[len(fields)-2] // TODO: this could panic and/or be totally incorrect

		retryAttempts, err := strconv.Atoi(retryAttemptsString)
		if err != nil {
			log.Printf("warn: crontab file line %d invalid: retry attempt value\n", lineCount)
			continue
		}

		retryDelayString := fields[len(fields)-1] // // TODO: this could panic and/or be totally incorrect

		retryDelay, err := strconv.Atoi(retryDelayString)
		if err != nil {
			log.Printf("warn: crontab file line %d invalid: retry delay value\n", lineCount)
			continue
		}

		// TODO: handle different problem cases, eg. where the schedule or command are empty or invalid (or leave it to the AddJob() ?)

		customCronJob := CustomCronJob{
			ID:            uuid.NewString(),
			Schedule:      schedule,
			Command:       command,
			RetryAttempts: retryAttempts,
			RetryDelay:    retryDelay,
			Topic:         topic,
			Client:        nil, // this is set later by ScheduleCustomCronJobs
		}

		customCronJobs = append(customCronJobs, customCronJob)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: failed reading through crontab file: %w", err)
	}

	return customCronJobs, nil
}

func ScheduleCustomCronJobs(customCronJobs []CustomCronJob, client *kgo.Client) *cron.Cron {
	cronParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronScheduler := cron.New(cron.WithParser(cronParser))

	for _, customCronJob := range customCronJobs {
		customCronJob.Client = client

		entryID, err := cronScheduler.AddJob(customCronJob.Schedule, customCronJob)
		if err != nil {
			log.Printf("error: failed to schedule cron job: %v\n", err)

			// metrics
			cronJobSchedulingFailures.Inc()

			continue
		}

		// metrics
		cronJobsScheduled.Inc()

		log.Printf("✅ cron job scheduled | entryID: %d | customCronJob: %v\n", entryID, customCronJob)
	}

	return cronScheduler
}

// func (c *cron.Cron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
// ^ to satisfy the Job interface that has a single method Run()
func (cj CustomCronJob) Run() {
	value := CustomCronJobValue{
		ID:            cj.ID,
		Schedule:      cj.Schedule,
		Command:       cj.Command,
		Topic:         cj.Topic,
		RetryAttempts: cj.RetryAttempts,
		RetryDelay:    cj.RetryDelay,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		log.Printf("error: failed to marshal cron job value to json: %v\n", err)
		return
	}

	// necessary to convert the ID to a byte slice to satisfy the kgo.Record type
	idByteSlice := []byte(cj.ID)

	record := &kgo.Record{
		Topic: cj.Topic,
		Key:   idByteSlice,
		Value: valueJSON,
	}

	// TODO: fix the context here to allow cancellation
	ctx := context.Background()

	cj.Client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		if err != nil {
			log.Printf("error: failed to produce record: %v\n", err)

			// metrics
			cronJobProducingFailures.Inc()

			return
		}

		// metrics
		cronJobsProduced.Inc()

		PrintProducedRecord(record)
	})

	// TODO: But how are errors handled from this Run() function...?
}
