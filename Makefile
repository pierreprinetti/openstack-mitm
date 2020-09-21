
test: unit-test system-test
.PHONY: test

unit-test:
	go test -v ./...
.PHONY: unit-test

system-test: osproxy:=$(shell mktemp)
system-test:
	go build -o $(osproxy) .
	./test/test $(osproxy)
	rm $(osproxy)
.PHONY: system-test
