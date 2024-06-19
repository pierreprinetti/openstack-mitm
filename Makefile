
osp-mitm: cmd/osp-mitm pkg/proxy
	go build ./cmd/osp-mitm

test: test-unit test-integration
.PHONY: test

test-unit:
	go test -v ./...
.PHONY: test-unit
