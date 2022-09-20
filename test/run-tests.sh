#!/usr/bin/env bash

go mod download
echo "Current dir - "
pwd
ls
echo "go.mod Before -"
cat go.mod
echo "go.sum Before -"
cat go.sum
cp go.mod go.mod.bak
cp go.sum go.sum.bak
go mod tidy
diff go.mod go.mod.bak
diff go.sum go.sum.bak
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
