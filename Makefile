.PHONY: proto build run clean

PROTO_FILES := $(shell find /proto -name '*.proto')
PROTO_GO_FILES := $(PROTO_FILES:.proto=.pb.go)

proto: $(PROTO_GO_FILES)

%.pb.go: %.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$<

build: proto
	go build -o bin/gateway cmd/gateway/main.go
	go build -o bin/bob cmd/bob/main.go
	go build -o bin/alice cmd/alice/main.go

run: build
	docker-compose up

clean:
	rm -f $(PROTO_GO_FILES)
	rm -rf bin
