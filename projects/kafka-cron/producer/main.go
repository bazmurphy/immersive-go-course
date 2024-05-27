package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	crontabFile := flag.String("path", "", "the path to the crontab file")
	flag.Parse()

	if *crontabFile == "" {
		flag.Usage()
		log.Fatalln("error: failed to provide the path to the crontab file")
	}

	log.Printf("opening crontab file...")

	file, err := os.Open(*crontabFile)
	if err != nil {
		log.Fatalf("error: failed to open crontab file: %s\n", err)
	}
	defer file.Close()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create a new parser with a seconds field
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	cronScheduler := cron.New(cron.WithParser(parser))

	log.Printf("new kafka client starting...")

	seeds := []string{"localhost:9092"}
	// TODO: pass seeds in as a flag/environment variable?

	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
	)
	if err != nil {
		log.Fatalf("error: failed to create new kafka client: %v\n", err)
	}
	defer client.Close()

	log.Println("reading crontab file...")

	scanner := bufio.NewScanner(file)

	var lineCount int

	for scanner.Scan() {
		lineCount++

		line := scanner.Text()

		// TODO: there are (as always) infinite possibilities of what could be wrong with each line...

		// skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 7 {
			log.Printf("warn: crontab file line %d invalid (expect a minimum of 7 fields)\n", lineCount)
			continue
		}
		// TODO: there is no guarantee just because we have 7+ fields that it is correct

		schedule := strings.Join(fields[:6], " ")
		command := strings.Join(fields[6:], " ")

		// TODO: handle different problem cases, eg. where the schedule or command are empty or invalid (or leave it to the AddJo)

		id := uuid.NewString()

		topic := "cron-topic"
		// TODO: should this topic should be passed in as a flag/environment variable?

		job := CustomCronJob{
			ID:          id,
			Schedule:    schedule,
			Command:     command,
			KafkaClient: client,
			Topic:       topic,
			Context:     ctx,
		}

		entryID, err := cronScheduler.AddJob(schedule, job)
		if err != nil {
			log.Printf("error: failed to add cronjob: %v\n", err)
			continue
		}

		log.Printf("cronjob scheduled with entry ID: %d\n", entryID)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error: failed reading through crontab file: %v\n", err)
	}

	log.Println("cron scheduler starting...")

	cronScheduler.Start()

	// wait for the termination signal
	<-signalChannel

	log.Println("cancelling the producer context...")

	// cancel the producer context
	cancel()

	log.Println("closing the kafka client...")

	client.Close()
	// TODO: with the defer client.Close() near the top is this really necessary? I want to control the order of shutdown...

	log.Println("cron scheduler stopping...")

	cronScheduler.Stop()

	log.Println("cron scheduler stopped")
}

type CustomCronJob struct {
	ID          string
	Schedule    string
	Command     string
	KafkaClient *kgo.Client
	Topic       string
	Context     context.Context
}

type CustomCronJobValue struct {
	ID       string `json:"id"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
}

// func (c *cron.Cron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
// ^ to satisfy the Job interface that has a single method Run()
func (cj CustomCronJob) Run() {
	log.Println("running cron job...")

	value := CustomCronJobValue{
		ID:       cj.ID,
		Schedule: cj.Schedule,
		Command:  cj.Command,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		log.Printf("error: failed to marshal cron job to json: %v\n", err)
		return
	}

	// necessary to convert the ID to a byte slice to satisfy the kgo.Record type
	idByteSlice := []byte(cj.ID)

	record := &kgo.Record{
		Topic: cj.Topic,
		Key:   idByteSlice,
		Value: valueJSON,
	}

	log.Println("kafka client producing new record...")

	cj.KafkaClient.Produce(cj.Context, record, func(_ *kgo.Record, err error) {
		if err != nil {
			log.Printf("error: failed to produce record: %v\n", err)
			return
		}
		log.Printf("produced record:\n\tkey:%s\n\tvalue:%s\n\ttopic:%s\n\tpartition:%d\n\toffset:%d\n\ttimestamp:%v", record.Key, record.Value, record.Topic, record.Partition, record.Offset, record.Timestamp)
	})

	// TODO: how are errors handled from this Run() function...
}
