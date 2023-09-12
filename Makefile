
test: test-unit test-integration
.PHONY: test

test-unit:
	go test -v ./...
.PHONY: test-unit

test-integration: osproxy:=$(shell mktemp)
test-integration:
	go build -o $(osproxy) .
	./test/test $(osproxy)
	rm $(osproxy)
.PHONY: system-test
