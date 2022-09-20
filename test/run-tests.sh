#!/usr/bin/env bash

echo "Current dir - "
pwd
ls
echo "go.mod Before -"
cat go.mod
cp go.mod go.mod.bak
go mod tidy
echo "go.mod Differences after -"
diff go.mod go.mod.bak

go mod download
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
