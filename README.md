# os-proxy

Proxies calls to the OpenStack API with a self-signed certificate.

All URLs in the OpenStack catalog are rewritten to point to the proxy itself, which will properly reverse proxy them to the original URL.

## Use locally

Download the binary for linux64 on this repository's [release page](https://github.com/shiftstack/os-proxy/releases) or build it with `go build .`.

**Required configuration:**
* **--remote-authurl**: URL of the remote OpenStack Keystone.
* **--proxy-url**: URL the proxy will be reachable at.

**Optional configuration:**
* **--remote-cacert**: path of the local PEM-encoded file containing the CA for the remote certificate.
* **--insecure**: skip TLS verification.

Example:
```shell
./os-proxy \
	--remote-authurl https://openstack.example.com:13000/v3 \
    --remote-cacert /var/openstack/cert.pem \
    --proxy-url https://localhost:15432'
```

## Deploy on the OpenStack cloud

The `proxy.sh` helper script deploys os-proxy to an OpenStack VM and attaches a floating IP for external connectivity.
The Ignition configuration injected in the VM triggers the download of a prebuilt `os-proxy` binary from Github.

Set `OS_PROXY` in the environment, and have the `openstack` client in $PATH.

**Required configuration:**
* **-a**: URL of the remote OpenStack Keystone.
* **-u**: URL the proxy will be reachable at.
* **-f**: Flavor of the proxy Nova instance.
* **-i**: Image of the proxy Nova instance.
* **-u**: Name or ID of the public network where to create the floating IP.

**Example:**
```shell
./proxy.sh \
	-a 'https://keystone.example.com:13000' \
	-u 'https://proxy.example.com:5443'     \
	-f 'm1.s2.medium'                       \
	-i 'rhcos'                              \
	-n 'external'
```

## Test

Run `make test`.

Requirements for the test:
* Bash v4+
* Go
* Netcat
* Jq
