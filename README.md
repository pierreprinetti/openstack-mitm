# openstack-mitm

Proxies calls to the OpenStack API with a self-signed certificate.

All URLs in the OpenStack catalog are rewritten to point to the proxy itself, which will properly reverse proxy them to the original URL.

## Use locally

Download the binary for linux64 on this repository's [release page](https://github.com/pierreprinetti/openstack-mitm/releases) or build it with `go build ./cmd/osp-mitm`.

**Required configuration:**
* **--remote-authurl**: URL of the remote OpenStack Keystone.
* **--proxy-url**: URL the proxy will be reachable at.

**Optional configuration:**
* **--remote-cacert**: path of the local PEM-encoded file containing the CA for the remote certificate.
* **--insecure**: skip TLS verification.

Example:
```shell
./osp-mitm \
	--remote-authurl https://openstack.example.com:13000/v3 \
	--remote-cacert /var/openstack/cert.pem \
	--proxy-url https://localhost:15432'
```
