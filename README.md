# openstack-proxy

`openstack-proxy` proxies calls to OpenStack, exposing the OpenStack API locally over http or https with a provided certificate.

All URLs in the OpenStack catalog are rewritten to point to the proxy itself, which will properly reverse-proxy them to the original URL.

## Use locally

Build with `go build ./cmd/openstack-proxy`.

`openstack-proxy` will parse a `clouds.yaml` file at the known locations, similar to what python-openstackclient does.

By default the server will listen on localhost on port 13000.

**Configuration:**
* `--url <url>`: the address this proxy will be reachable at. Default: `http://locahost:13000`
* `--cert <path>`: path of the local PEM-encoded HTTPS certificate file. Mandatory if the scheme of --url is HTTPS
* `--key <path>`: path of the local PEM-encoded HTTPS certificate key file. Mandatory if the scheme of --url is HTTPS
* `-o <path>`: location where to write a new `clouds.yaml` that points to the openstack-proxy instance

## Examples

### Local server

```bash
export OS_CLOUD=openstack
./openstack-proxy -o proxied-clouds.yaml
```
```bash
export OS_CLIENT_CONFIG_FILE=./proxied-clouds.yaml
openstack server list
```

### On the network, with HTTPS

```bash
./openstack-proxy \
	--url https://myserver.example.com:13000 \
	--cert /var/run/osp-cert.pem \
	--key /var/run/osp-key.pem' \
	-o proxied-clouds.yaml
```

## Containerized

An image of `openstack-proxy` is available on `quay.io/pierreprinetti/openstack-proxy`.

Here is an example of how to use it. This example uses the clouds.yaml in the
default location, edit as necessary:

```bash
export OS_CLOUD=openstack # Name of the target cloud in clouds.yaml
export OS_CLIENT_CONFIG_FILE="$(mktemp -d)/clouds.yaml" # Where to write the generated clouds.yaml
podman run \
        --name openstack-proxy \
        -d \
        --rm \
        -v ${XDG_CONFIG_HOME:-${HOME}/.config}/openstack/clouds.yaml:/config/clouds.yaml:Z,ro \
        -v $(dirname "$OS_CLIENT_CONFIG_FILE"):/out:z \
        --env OS_CLOUD=$OS_CLOUD \
        -p 13000:13000 \
        quay.io/pierreprinetti/openstack-proxy -o "/out/$(basename "$OS_CLIENT_CONFIG_FILE")"
```

Logs can be inspected in a separate terminal:

```bash
podman container logs --follow openstack-proxy
```

To terminate the container:

```bash
podman container stop openstack-proxy
```
