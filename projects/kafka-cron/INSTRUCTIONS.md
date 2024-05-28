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

### 1. Use Makefile
- from the base directory in a new shell `make clusters-up` (to bring up the clusters) and `make clusters-down` (to bring down the clusters)
- from the base directory in a new shell `make clients-up` (to bring up the project) and `make clients-down` (to bring down the project)
  #### What happens:
    1. the `admin` creates the `cron-topic` on both clusters and shows `metadata` about both clusters
    2. the `producer` reads the `customcrontab` file and starts a long running cron scheduler that produces cronjobs to either `cluster-a` or `cluster-b` (depending on the crontab line) to the `cron-topic`s
    3. the `consumer`s start consuming cronjobs from either `cluster-a` or `cluster-b` `cron-topic` (depending on their configuration) and execute the commands from the cronjobs