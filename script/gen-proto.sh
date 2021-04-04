#!/bin/bash
# GOOGLE_APIS=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
PROTO_DIR=./api/proto
OUT_DIR=./pkg/api/user

protoc -I $GOPATH/src \
    -I vendor/github.com/grpc-ecosystem/grpc-gateway/ \
    -I vendor/ \
    -I $PROTO_DIR/ \
    --grpc-gateway_out=$OUT_DIR \
    --swagger_out=$OUT_DIR/third_party/OpenAPI/ \
    --go_out=plugins=grpc:$OUT_DIR \
    $PROTO_DIR/*.proto

statik -m -f -src $OUT_DIR/third_party/OpenAPI/ --dest $OUT_DIR