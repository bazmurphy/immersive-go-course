# Makefile

.PHONY: clusters-up clusters-down clients-up clients-down

clusters-up:
	docker compose -f docker-compose-clusters.yml up

clusters-down:
	docker compose -f docker-compose-clusters.yml down

clients-up:
	docker compose -f docker-compose-clients.yml up --build

clients-down:
	docker compose -f docker-compose-clients.yml down
