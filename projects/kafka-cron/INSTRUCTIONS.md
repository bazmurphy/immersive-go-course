# Instructions

## 1. Bring up the Kafka Cluster
- from the base directory `docker compose -f docker-compose-cluster.yml up`

## 2. Bring up the Clients: Admin, Producer, Consumers
- from the base directory `docker compose -f docker-compose-clients.yml up`
  ### What happens:
  1. the `admin` creates the `cron-topic` and shows `metadata`
  2. the `producer` reads the `customcrontab` file and starts a long running cron scheduler that produces cronjobs to the `cron-topic`
  3. the `consumer`s start consuming cronjobs from the `cron-topic` and execute the commands from the cronjobs