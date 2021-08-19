#!/usr/bin/env bash

go test ./...
if [[ "$?" != "0" ]]; then
  echo "go test ./... failed..."
  exit 1
fi
