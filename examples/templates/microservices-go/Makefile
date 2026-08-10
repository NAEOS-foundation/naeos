.PHONY: run test generate proto

run:
	go run ./cmd/api

test:
	go test -race ./...

generate:
	naeos generate spec.yaml --config naeos.yaml

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/order/v1/order.proto
