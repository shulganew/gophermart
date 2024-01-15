
#Migrations

.PHONY: db-init
db-init:
	docker run --rm \
    	-v $(realpath ./db/migrations):/migrations \
    	migrate/migrate:v4.16.2 \
        	create \
        	-dir /migrations \
        	-ext .sql \
        	-seq -digits 3 \
        	init


.PHONY: pg
pg: 
	docker run --rm \
		--name=marketdb_v1 \
		-v $(abspath ./db/init/):/docker-entrypoint-initdb.d \
		-e POSTGRES_PASSWORD="postgres" \
		-d \
		-p 5432:5432 \
		postgres:15.3


.PHONY: pg-stop
pg-stop:
	docker stop marketdb_v1

.PHONY: clean-data
clean-data:
	sudo rm -rf ./db/data/

.PHONY: pg-up
pg-up:
	docker run --rm \
    -v $(realpath ./db/migrations):/migrations \
    migrate/migrate:v4.16.2 \
        -path=/migrations \
        -database postgres://market:1@172.17.0.2:5432/market?sslmode=disable \
        up


#Linter 

GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.55.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint 
