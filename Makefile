GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: lintall
lintall: fmt lint

.PHONY:
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	@gofmt -d ${GOFILES}; \
	if [ -n "$$(gofmt -l ${GOFILES})" ]; then \
	  echo "Please run 'make dofmt'" && exit 1; \
	fi

.PHONY: travis-lint
travis-lint:
	yamllint .travis.yml

.PHONY: gosec
gosec:
	gosec -quiet --exclude=G104 ./...

.PHONY: scan-nancy
scan-nancy:
	go mod tidy
	go list -json -m all | nancy sleuth --skip-update-check
