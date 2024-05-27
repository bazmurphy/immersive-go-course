# Instructions

## Part 1

### 1. Bring up the Kafka Cluster
- from the base directory `docker compose -f docker-compose-cluster.yml up`

### 2. Bring up the Clients: Admin, Producer, Consumers
- from the base directory `docker compose -f docker-compose-clients.yml up`
  #### What happens:
  1. the `admin` creates the `cron-topic` on the cluster and shows `metadata` about the cluster
  2. the `producer` reads the `customcrontab` file and starts a long running cron scheduler that produces cronjobs to the `cron-topic`
  3. the `consumer`s start consuming cronjobs from the `cron-topic` and execute the commands from the cronjobs

## Part 2

### 1. Bring up both Kafka Clusters - `cluster-a` and `cluster-b`
- from the base directory `docker compose -f docker-compose-cluster-a.yml up`
- from the base directory `docker compose -f docker-compose-cluster-b.yml up`

### 2. Bring up the Clients: Admin, Producer, Consumers
- from the base directory `docker compose -f docker-compose-clients.yml up`
  #### What happens:
    1. the `admin` creates the `cron-topic` on both clusters and shows `metadata` about both clusters
    2. the `producer` reads the `customcrontab` file and starts a long running cron scheduler that produces cronjobs to either `cluster-a` or `cluster-b` (depending on the crontab line) to the `cron-topic`s
    3. the `consumer`s start consuming cronjobs from either `cluster-a` or `cluster-b` `cron-topic` (depending on their configuration) and execute the commands from the cronjobs