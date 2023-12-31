#===================#
#== Env Variables ==#
#===================#

DOCKER_COMPOSE_FILE ?= docker-compose.dev.yml

#=========================#
#== DATABASE MANAGEMENT ==#
#=========================#

up-db:
	docker compose -f ${DOCKER_COMPOSE_FILE} up

build-db:
	docker compose -f ${DOCKER_COMPOSE_FILE} up -d --build

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

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github/leoflalv/bank-api/db/sqlc Store
