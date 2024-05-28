// go run . -path=customcrontab -cluster-a-seeds=localhost:9092,localhost:9093,localhost:9094 -cluster-b-seeds=localhost:9095,localhost:9096,localhost:9097 -topic=cron-topic

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	crontabFilePathFlag string
	clusterASeedsFlag   string
	clusterBSeedsFlag   string
	topicFlag           string
)

func main() {
	flag.StringVar(&crontabFilePathFlag, "path", "", "the path to the crontab file")
	flag.StringVar(&clusterASeedsFlag, "cluster-a-seeds", "", "the kafka broker addresses of cluster a")
	flag.StringVar(&clusterBSeedsFlag, "cluster-b-seeds", "", "the kafka broker addresses of cluster b")
	flag.StringVar(&topicFlag, "topic", "", "the name of the topic")

	flag.Parse()
	// log.Println("DEBUG | crontabFilePathFlag:", crontabFilePathFlag, "clusterASeedsFlag:", clusterASeedsFlag, "clusterBSeedsFlag:", clusterBSeedsFlag, "topicFlag:", topicFlag)

	// TODO: handle the flag errors more specifically
	if crontabFilePathFlag == "" || clusterASeedsFlag == "" || clusterBSeedsFlag == "" || topicFlag == "" {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// --------------------------------------------

	log.Println("establishing cluster connections...")

	clientClusterA, clientClusterB, err := ClusterConnections()
	if err != nil {
		log.Fatalf("error: failed to establish cluster connections: %v", err)
	}
	defer clientClusterA.Close()
	defer clientClusterB.Close()

	log.Println("cluster connections established")

	// --------------------------------------------

	log.Println("opening crontab file...")

	cronTabFile, err := os.Open(crontabFilePathFlag)
	if err != nil {
		log.Fatalf("error: failed to open crontab file: %v\n", err)
	}
	defer cronTabFile.Close()

	// --------------------------------------------

	log.Println("parsing crontab file...")

	customCronJobs, err := ParseCronTabFile(cronTabFile)
	if err != nil {
		log.Fatalf("error: failed to parse crontab file: %v\n", err)
	}

	log.Printf("parsed %d cron jobs\n", len(customCronJobs))

	// --------------------------------------------

	log.Println("scheduling cron jobs...")

	cronScheduler := ScheduleCustomCronJobs(customCronJobs, clientClusterA, clientClusterB)

	log.Printf("scheduled %d cron jobs\n", len(cronScheduler.Entries()))

	// --------------------------------------------

	log.Println("cron scheduler starting...")

	cronScheduler.Start()

	<-signalChannel
	log.Println("received termination signal, cron scheduler stopping...")

	cronScheduler.Stop()

	log.Println("cron scheduler stopped")

	// --------------------------------------------
}

func ClusterConnections() (*kgo.Client, *kgo.Client, error) {
	clusterASeeds := strings.Split(clusterASeedsFlag, ",")
	clusterBSeeds := strings.Split(clusterBSeedsFlag, ",")

	clientClusterA, err := kgo.NewClient(
		kgo.SeedBrokers(clusterASeeds...),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error: failed to create new cluster-a kafka client: %w", err)
	}
	// defer clientClusterA.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = clientClusterA.Ping(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error: failed to ping cluster-a: %w", err)
	}

	clientClusterB, err := kgo.NewClient(
		kgo.SeedBrokers(clusterBSeeds...),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error: failed create new cluster-b kafka client: %w", err)
	}
	// defer clientClusterB.Close()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = clientClusterB.Ping(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error: failed to ping cluster-b: %w", err)
	}

	return clientClusterA, clientClusterB, nil
}

type CustomCronJob struct {
	ID            string
	Schedule      string
	Command       string
	Cluster       string
	RetryAttempts int
	Topic         string
	Client        *kgo.Client
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

		cluster := fields[len(fields)-2] // TODO: this could panic and/or be totally incorrect

		if cluster != "cluster-a" && cluster != "cluster-b" {
			log.Printf("warn: crontab file line %d invalid: cluster name\n", lineCount)
			continue
		}

		retryAttemptsString := fields[len(fields)-1] // TODO: this could panic and/or be totally incorrect

		retryAttempts, err := strconv.Atoi(retryAttemptsString)
		if err != nil {
			log.Printf("warn: crontab file line %d invalid: retry attempt value\n", lineCount)
			continue
		}

		// TODO: handle different problem cases, eg. where the schedule or command are empty or invalid (or leave it to the AddJob() ?)

		customCronJob := CustomCronJob{
			ID:            uuid.NewString(),
			Schedule:      schedule,
			Command:       command,
			Cluster:       cluster,
			RetryAttempts: retryAttempts,
			Topic:         topicFlag,
			Client:        nil, // this is provided later
		}

		customCronJobs = append(customCronJobs, customCronJob)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error: failed reading through crontab file: %w", err)
	}

	return customCronJobs, nil
}

func ScheduleCustomCronJobs(customCronJobs []CustomCronJob, clientClusterA *kgo.Client, clientClusterB *kgo.Client) *cron.Cron {
	cronParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronScheduler := cron.New(cron.WithParser(cronParser))

	for _, customCronJob := range customCronJobs {
		switch customCronJob.Cluster {
		case "cluster-a":
			customCronJob.Client = clientClusterA
		case "cluster-b":
			customCronJob.Client = clientClusterB
		}

		entryID, err := cronScheduler.AddJob(customCronJob.Schedule, customCronJob)
		if err != nil {
			log.Printf("error: failed to schedule cron job: %v\n", err)
			continue
		}
		log.Printf("cron job scheduled | entryID: %.2d | customCronJob: %v\n", entryID, customCronJob)
	}

	return cronScheduler
}

type CustomCronJobValue struct {
	ID            string `json:"id"`
	Schedule      string `json:"schedule"`
	Command       string `json:"command"`
	Cluster       string `json:"cluster"`
	RetryAttempts int    `json:"retry_attempts"`
}

// func (c *cron.Cron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
// ^ to satisfy the Job interface that has a single method Run()
func (cj CustomCronJob) Run() {
	value := CustomCronJobValue{
		ID:            cj.ID,
		Cluster:       cj.Cluster,
		Schedule:      cj.Schedule,
		Command:       cj.Command,
		RetryAttempts: cj.RetryAttempts,
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

	log.Println("producing new record...")

	// TODO: fix the context here
	cj.Client.Produce(context.Background(), record, func(_ *kgo.Record, err error) {
		if err != nil {
			log.Printf("error: failed to produce record: %v\n", err)
			return
		}
		printProducedRecord(record, cj.Cluster)
	})

	// TODO: But how are errors handled from this Run() function...?
}

func printProducedRecord(cj *kgo.Record, cluster string) {
	red := "\x1b[31m"
	green := "\x1b[32m"
	yellow := "\x1b[33m"
	blue := "\x1b[34m"
	magenta := "\x1b[35m"
	// cyan := "\x1b[36m"
	reset := "\x1b[0m"

	log.Printf(
		"%sProduced Record%s:\n\t%sCluster%s: %s%s%s\n\t%sTopic%s: %s%s%s\n\t%sPartition%s: %s%d%s\n\t%sOffset%s: %s%d%s\n\t%sTimestamp%s: %s%v%s\n\t%sKey%s: %s%s%s\n\t%sValue%s: %s%s%s\n",
		yellow, reset,
		blue, reset, red, cluster, reset,
		blue, reset, red, cj.Topic, reset,
		blue, reset, magenta, cj.Partition, reset,
		blue, reset, magenta, cj.Offset, reset,
		blue, reset, magenta, cj.Timestamp, reset,
		blue, reset, green, string(cj.Key), reset,
		blue, reset, green, string(cj.Value), reset,
	)
}
