#!/usr/bin/env bash

GRPC_GW_PATH="$(pwd)/third_party/googleapis"

mkdir -p third_party
[ ! -d "${GRPC_GW_PATH}" ] && git clone https://github.com/googleapis/googleapis.git third_party/googleapis

# generate the gRPC code
protoc -I. -I"${GRPC_GW_PATH}" --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  exchange.proto

# generate the JSON interface code
protoc -I. -I"${GRPC_GW_PATH}"  --grpc-gateway_out=paths=source_relative,logtostderr=true:. \
  exchange.proto

# generate the swagger definitions
protoc -I. -I"${GRPC_GW_PATH}" --swagger_out=json_names_for_fields=true,simple_operation_ids=true:./swagger \
  exchange.proto

# merge the swagger code into one file
go run swagger/main.go swagger > ../../static/swagger/api.swagger.json
