# os-proxy

Proxies calls to the OpenStack API with a self-signed certificate.

All URLs in the OpenStack catalog are rewritten to point to the proxy itself, which will properly reverse proxy them to the original URL.

**Required configuration:**
* **-a**: URL of the remote OpenStack Keystone.
* **-u**: URL the proxy will be reachable at.
* **-f**: Flavor of the proxy Nova instance.
* **-i**: Image of the proxy Nova instance.
* **-u**: Name or ID of the public network where to create the floating IP.

**Example:**
```
./proxy.sh \
	-a 'https://keystone.example.com:13000' \
	-u 'https://proxy.example.com:5443'     \
	-f 'm1.s2.medium'                       \
	-i 'rhcos'                              \
	-n 'external'
```
