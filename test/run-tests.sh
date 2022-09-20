#!/usr/bin/env bash

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin
${GOPATH}/bin/golangci-lint run ./...

if [[ "$?" != "0" ]]; then
  echo "golangci-lint run ./... failed..."
  exit 1
fi

go test ./...
if [[ "$?" != "0" ]]; then
  echo "go test ./... failed..."
  exit 1
fi
