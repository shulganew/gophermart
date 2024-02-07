#!/bin/bash
export PATH="/home/igor/Desktop/code/go-autotests-0.10.2/bin:$PATH"
#./cmd/accrual/accrual_linux_amd64 &

function check(){
	res=""
	if [ $2 -ne 0 ]; then res=$(echo "$1: {$res} Error! $2"); echo "ERROR!  Iter:" $res;exit 1; else res=$(echo "$1: ${res} PASS "); fi
	echo "Iter:" $res
}
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://postgres:postgres@postgres/praktikum" goose -dir ./migrations  up
PGPASSWORD=postgres psql -h postgres -U postgres -d praktikum -c "truncate TABLE users cascade"

go build -o ./cmd/gophermart/gophermart  ./cmd/gophermart/main.go
go vet -vettool=$(which statictest) ./...
check S $? 

          gophermarttest \
            -test.v -test.run=^TestGophermart$ \
            -gophermart-binary-path=cmd/gophermart/gophermart \
            -gophermart-host=localhost \
            -gophermart-port=8088 \
            -gophermart-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable" \
            -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
            -accrual-host=localhost \
            -accrual-port=$(random unused-port) \
            -accrual-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"
check 1 $? make 