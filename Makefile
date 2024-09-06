
openstack-proxy: cmd/openstack-proxy pkg/proxy
	go build ./cmd/openstack-proxy

test: test-unit test-integration
.PHONY: test

test-unit:
	go test -v ./...
.PHONY: test-unit
