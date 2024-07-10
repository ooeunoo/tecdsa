.PHONY: proto build run clean reset

PROTO_DIR := proto
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')
PROTO_GO_FILES := $(PROTO_FILES:.proto=.pb.go)
PROTO_GRPC_GO_FILES := $(PROTO_FILES:.proto=_grpc.pb.go)

proto: $(PROTO_GO_FILES) $(PROTO_GRPC_GO_FILES)

%.pb.go %_grpc.pb.go: %.proto
	protoc -I=$(PROTO_DIR) --go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		$<

build: proto
	go build -o bin/gateway cmd/gateway/main.go
	go build -o bin/bob cmd/bob/main.go
	go build -o bin/alice cmd/alice/main.go

run: build
	docker-compose up

clean:
	find $(PROTO_DIR) -name "*.pb.go" -type f -delete
	find $(PROTO_DIR) -name "*_grpc.pb.go" -type f -delete
	rm -rf bin

reset:
	find cmd -name "data" -type d -exec rm -rf {} +
