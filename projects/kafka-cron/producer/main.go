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

	// log.Println("DEBUG | crontabFilePathFlag", crontabFilePathFlag, "clusterASeedsFlag", clusterASeedsFlag, "clusterBSeedsFlag", clusterBSeedsFlag, "topicFlag", topicFlag, "partitionsFlag")

	if crontabFilePathFlag == "" || clusterASeedsFlag == "" || clusterBSeedsFlag == "" || topicFlag == "" {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	log.Printf("opening crontab file...")

	file, err := os.Open(crontabFilePathFlag)
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

	clusterASeeds := strings.Split(clusterASeedsFlag, ",")
	clusterBSeeds := strings.Split(clusterBSeedsFlag, ",")

	log.Printf("new kafka clients starting (one for cluster-a, one for cluster-b)...")

	clientClusterA, err := kgo.NewClient(
		kgo.SeedBrokers(clusterASeeds...),
	)
	if err != nil {
		log.Fatalf("error: failed to create new cluster-a kafka client: %v\n", err)
	}
	defer clientClusterA.Close()

	clientClusterB, err := kgo.NewClient(
		kgo.SeedBrokers(clusterBSeeds...),
	)
	if err != nil {
		log.Fatalf("error: failed to create new cluster-b kafka client: %v\n", err)
	}
	defer clientClusterB.Close()

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

		if len(fields) < 8 {
			log.Printf("warn: crontab file line %d invalid (expect a minimum of 8 fields)\n", lineCount)
			continue
		}
		// TODO: there is no guarantee just because we have 8+ fields that it is correct

		schedule := strings.Join(fields[:6], " ")

		command := strings.Join(fields[6:len(fields)-1], " ")

		cluster := fields[len(fields)-1]
		// TODO: this could panic and/or be totally incorrect

		if cluster != "cluster-a" && cluster != "cluster-b" {
			log.Printf("warn: crontab file line %d invalid cluster name\n", lineCount)
			continue
		}

		// TODO: handle different problem cases, eg. where the schedule or command are empty or invalid (or leave it to the AddJo)

		id := uuid.NewString()

		topic := topicFlag

		var client *kgo.Client

		switch cluster {
		case "cluster-a":
			client = clientClusterA
		case "cluster-b":
			client = clientClusterB
		}

		job := CustomCronJob{
			ID:          id,
			Schedule:    schedule,
			Command:     command,
			Cluster:     cluster,
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

	log.Println("cron scheduler stopping...")

	cronScheduler.Stop()

	log.Println("cron scheduler stopped")
}

type CustomCronJob struct {
	ID          string
	Schedule    string
	Command     string
	Cluster     string
	KafkaClient *kgo.Client
	Topic       string
	Context     context.Context
}

type CustomCronJobValue struct {
	Cluster  string `json:"cluster"`
	ID       string `json:"id"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
}

// func (c *cron.Cron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
// ^ to satisfy the Job interface that has a single method Run()
func (cj CustomCronJob) Run() {
	log.Println("running cron job...")

	value := CustomCronJobValue{
		Cluster:  cj.Cluster,
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
		log.Printf("produced record:\n\tcluster:%s\n\ttopic:%s\n\tpartition:%d\n\toffset:%d\n\ttimestamp:%v\n\tkey:%s\n\tvalue:%s\n", cj.Cluster, record.Topic, record.Partition, record.Offset, record.Timestamp, record.Key, record.Value)
	})

	// TODO: how are errors handled from this Run() function...
}
