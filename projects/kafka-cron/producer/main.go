// go run . -path=customcrontab

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	crontabFilePathFlag string
	seedsFlag           string
)

func main() {
	flag.StringVar(&crontabFilePathFlag, "path", "", "the path to the crontab file")
	flag.StringVar(&seedsFlag, "seeds", "kafka1:19092", "the kafka broker addresses")

	flag.Parse()
	// log.Println("DEBUG | crontabFilePathFlag:", crontabFilePathFlag, "seedsFlag:", seedsFlag)

	// TODO: handle the flag errors more specifically
	if crontabFilePathFlag == "" || seedsFlag == "" {
		flag.Usage()
		log.Fatalf("error: missing or invalid flag values")
	}

	// --------------------------------------------

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// --------------------------------------------

	log.Println("establishing cluster connection...")

	client, err := ProducerClient()
	if err != nil {
		log.Fatalf("error: failed to establish cluster connection: %v", err)
	}
	defer client.Close()

	log.Println("cluster connection established")

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

	cronScheduler := ScheduleCustomCronJobs(customCronJobs, client)

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
