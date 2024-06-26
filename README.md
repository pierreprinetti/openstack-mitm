# openstack-mitm

Proxies calls to the OpenStack API with a self-signed certificate.

All URLs in the OpenStack catalog are rewritten to point to the proxy itself, which will properly reverse proxy them to the original URL.

## Use locally

Build with `go build ./cmd/osp-mitm`.

`osp-mitm` will parse a `clouds.yaml` file at the known locations, similar to what python-openstackclient does.

By default the server will listen on localhost on port 13000.

**Configuration:**
* **--url**: URL osp-mitm will be reachable at. Default: `http://locahost:13000`
* **--cert**: path of the local PEM-encoded HTTPS certificate file. Mandatory if the scheme of --url is HTTPS.
* **--key**: path of the local PEM-encoded HTTPS certificate key file. Mandatory if the scheme of --url is HTTPS.
* **-o**: If provided, a new clouds.yaml that points to osp-mitm is created at that location.

## Examples

Local server:
```shell
export OS_CLOUD=openstack
./osp-mitm -o mitm-clouds.yaml
```
```shell
export OS_CLIENT_CONFIG_FILE=./mitm-clouds.yaml
openstack server list
```

Exposing osp-mitm on the network, with HTTPS:

```shell
./osp-mitm \
	--url https://myserver.example.com:13000 \
	--cert /var/run/osp-cert.pem \
	--key /var/run/osp-key.pem' \
	-o mitm-clouds.yaml
```
