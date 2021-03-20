
.PHONY: build

build:
	go build ./cmd/yarbit/

cleandb:
	rm -rf data/
