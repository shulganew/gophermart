

#Migrations

.PHONY: pg

pg: 
	docker run --rm \
		--name=marketdb_v2 \
		-e POSTGRES_PASSWORD="postgres" \
		-d \
		-p 5432:5432 \
		postgres:15.3
	sleep 1

	PGPASSWORD=postgres psql -v ON_ERROR_STOP=1 -h postgres -U postgres \
		-c "CREATE USER market WITH ENCRYPTED PASSWORD '1';" \
		-c "CREATE USER praktikum WITH ENCRYPTED PASSWORD 'praktikum';" \
		-c "CREATE DATABASE market;"  \
		-c "CREATE DATABASE praktikum;" \
		-c "GRANT ALL PRIVILEGES ON DATABASE market TO market;" \
		-c "GRANT ALL PRIVILEGES ON DATABASE praktikum TO market;" \
		-c "ALTER DATABASE market OWNER TO market;" \
		-c "ALTER DATABASE praktikum OWNER TO market;"


.PHONY: pg-stop
pg-stop:
	docker stop marketdb_v2




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
