#===================#
#== Env Variables ==#
#===================#
DOCKER_COMPOSE_FILE ?= docker-compose.dev.yml

#=========================#
#== DATABASE MANAGEMENT ==#
#=========================#

up-db:
	docker compose -f ${DOCKER_COMPOSE_FILE} up

shell-db:
	docker compose -f ${DOCKER_COMPOSE_FILE} exec db psql -U postgres -d postgres


#========================#
#== DATABASE MIGRATION ==#
#========================#

migrate-up:
	docker compose -f ${DOCKER_COMPOSE_FILE} --profile tools run --rm migrate up

migrate-down:
	docker compose -f ${DOCKER_COMPOSE_FILE} --profile tools run --rm migrate down 1

migrate-create:
	docker compose -f ${DOCKER_COMPOSE_FILE} --profile tools run --rm migrate create -ext sql -dir /migrations $(name)

#=================#
#== GO COMMANDS ==#
#=================#

test:
	go test -v -cover ./...
