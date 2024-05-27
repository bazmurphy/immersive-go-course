# Instructions

## 1. Bring up the Kafka Cluster
- from the base directory `docker compose -f zk-single-kafka-multiple.yml up`

## 2. Create the Kafka Topic
- from the `admin` directory `go run .`
  - to create the topic

## 3. Producer
- from the `producer` directory `go run . --path=customcrontab` 
  -  to read the `customcrontab` file and run a cron scheduler that produces cronjobs to the topic

## 4. Consumer
- from the `consumer` directory `go run .` 
  - to start consuming cronjobs from the topic and executing the cronjobs
