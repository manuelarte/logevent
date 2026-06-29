.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: fmt
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test -v ./...
