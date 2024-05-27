# Makefile

.PHONY: up down

up:
	docker compose -f docker-compose-cluster-a.yml up -d
	./wait-for-it.sh localhost:9092 --timeout=60
	docker compose -f docker-compose-cluster-b.yml up -d
	./wait-for-it.sh localhost:9094 --timeout=60
	docker compose -f docker-compose-clients.yml up

down:
	docker compose -f docker-compose-clients.yml down
	docker compose -f docker-compose-cluster-b.yml down
	docker compose -f docker-compose-cluster-a.yml down
