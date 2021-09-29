#!/usr/bin/env bash

 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.0
 golangci-lint run ./...

if [[ "$?" != "0" ]]; then
  echo "golangci-lint run ./... failed..."
  exit 1
fi

go test ./...
if [[ "$?" != "0" ]]; then
  echo "go test ./... failed..."
  exit 1
fi
