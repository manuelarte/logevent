.PHONY: fmt
fmt:
	golangci-lint fmt ./...

.PHONY: lint
lint: fmt
	golangci-lint custom -v
	./custom-gcl run --fix ./...

.PHONY: test
test:
	go test -v ./...
