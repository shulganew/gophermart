#!/bin/bash
export PATH="/home/igor/Desktop/code/go-autotests-0.10.2/bin:$PATH"


function check(){
	res=""
	if [ $2 -ne 0 ]; then res=$(echo "$1: {$res} Error! $2"); echo "ERROR!  Iter:" $res;exit 1; else res=$(echo "$1: ${res} PASS "); fi
	echo "Iter:" $res
}

go build -o ./cmd/gophermart ./cmd/gophermart/main.go
go vet -vettool=$(which statictest) ./...
check S $? 