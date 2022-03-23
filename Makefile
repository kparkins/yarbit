
.PHONY: build

build:
	go build ./cmd/yarbit/

build-race:
	go build -race ./cmd/yarbit

ec2:
	GOOS=linux GOARCH=amd64 go build ./cmd/yarbit/ 

debug:
	go build -gcflags="all=-N -l" ./cmd/yarbit

run: build
	./yarbit run --datadir=data --port=8080

status:
	curl -s 127.0.0.1:8080/node/status | jq

balances:
	curl -s 127.0.0.1:8080/balances/list | jq

addtx:
	curl -s -X POST 127.0.0.1:8080/tx/add -H 'Content-Type: application/json' -d '{"from": "andrej", "to": "kyle", "value": 1, "data": ""}' | jq

cleandb:
	rm -rf data/
