# Makefile

.PHONY: kafka-up kafka-down cron-up cron-down

kafka-up:
	docker compose -f docker-compose-kafka.yml up

kafka-down:
	docker compose -f docker-compose-kafka.yml down

cron-up:
	docker compose -f docker-compose-cron.yml up --build

cron-down:
	docker compose -f docker-compose-cron.yml down
