GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: lintall
lintall: fmt lint

.PHONY:
lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.0
	golangci-lint run ./...
	$(MAKE) deps

.PHONY: fmt
fmt:
	@gofmt -d ${GOFILES}; \
	if [ -n "$$(gofmt -l ${GOFILES})" ]; then \
	  echo "Please run 'make dofmt'" && exit 1; \
	fi

.PHONY: travis-lint
travis-lint:
	yamllint .travis.yml

.PHONY: deps
	go mod tidy

