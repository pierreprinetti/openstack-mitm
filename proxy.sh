#!/usr/bin/env bash

# Copyright 2020 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -Eeuo pipefail

declare -r \
	osproxy_url='https://github.com/shiftstack/os-proxy/releases/download/v1.0.1/os-proxy' \
	osproxy_sha512='d4a9210091e4d1ed4c697762ac5ed59625c97dbdf3ce58cc4bbd7f3821190f482e2464558fbd08ea737744a7cc496e9b6db4381c3941b8fb1c864d1bec35113f'

print_help() {
	echo -e "github.com/shiftstack/os-proxy"
	echo -e "Proxy calls to the OpenStack API"
	echo
	echo -e "Required configuration:"
	echo -e "\t-a\tURL of the remote OpenStack Keystone."
	echo -e "\t-u\tURL the proxy will be reachable at."
	echo -e "\t-f\tFlavor of the proxy Nova instance."
	echo -e "\t-i\tImage of the proxy Nova instance."
	echo -e "\t-u\tName or ID of the public network where to create the floating IP."
	echo
	echo -e "Example:"
	echo -e "\t${0} \\"
	echo -e "\t\t-a 'https://keystone.example.com:13000' \\"
	echo -e "\t\t-u 'https://proxy.example.com:5443'     \\"
	echo -e "\t\t-f 'm1.s2.medium'                       \\"
	echo -e "\t\t-i 'rhcos'                              \\"
	echo -e "\t\t-n 'external'"
}

declare \
	auth_url=''                 \
	proxy_url=''                \
	server_flavor=''            \
	server_image=''             \
	external_network='external'
while getopts a:u:f:i:n:h opt; do
	case "$opt" in
		a) auth_url="$OPTARG"         ;;
		u) proxy_url="$OPTARG"        ;;
		f) server_flavor="$OPTARG"    ;;
		i) server_image="$OPTARG"     ;;
		n) external_network="$OPTARG" ;;
		h) print_help; exit 0         ;;
		*) exit 1                     ;;
	esac
done
shift "$((OPTIND-1))"
readonly \
	auth_url \
	proxy_url \
	server_flavor \
	server_image \
	external_network

declare -r \
	name='osproxy' \
	port="${proxy_url##*:}"
declare \
	sg_id=''      \
	network_id='' \
	subnet_id=''  \
	router_id=''  \
	port_id=''    \
	server_id=''  \
	fip_id=''

{
	re='^[0-9]+$'
	if ! [[ $port =~ $re ]] ; then
		>&2 echo "The proxy URL ('-u') should have a port number."
		exit 1
	fi

	if (( $port < 1024 )); then
		>&2 echo "The proxy URL ('-u') port number should be higher than 1024."
		exit 1
	fi
}

cleanup() {
	>&2 echo
	>&2 echo
	>&2 echo 'Starting the cleanup...'
	if [ -f "$ignition" ]; then
		rm  "$ignition" || >&2 echo "Failed deleting temporary file '${ignition}'"
	fi
	if [ -n "$fip_id" ]; then
		openstack floating ip delete "$fip_id" || >&2 echo 'Failed deleting FIP'
	fi
	if [ -n "$server_id" ]; then
		openstack server delete "$server_id" || >&2 echo 'Failed deleting server'
	fi
	if [ -n "$port_id" ]; then
		openstack port delete "$port_id" || >&2 echo 'Failed deleting port'
	fi
	if [ -n "$router_id" ]; then
		openstack router remove subnet "$router_id" "$subnet_id" || >&2 echo 'Failed removing subnet from router'
		openstack router delete "$router_id" || >&2 echo 'Failed deleting router'
	fi
	if [ -n "$subnet_id" ]; then
		openstack subnet delete "$subnet_id" || >&2 echo "Failed deleting subnet"
	fi
	if [ -n "$network_id" ]; then
		openstack network delete "$network_id" || >&2 echo 'Failed deleting network'
	fi
	if [ -n "$sg_id" ]; then
		openstack security group delete "$sg_id" || >&2 echo 'Failed deleting security group'
	fi
	>&2 echo 'Cleanup done.'
}

trap cleanup EXIT

declare -r ignition="$(mktemp)"
cat > "$ignition" <<EOF
{
  "ignition": { "version": "3.1.0" },
  "passwd": {
    "users": [
      {
        "name": "osproxy"
      }
    ]
  },
  "storage": {
    "files": [{
      "path": "/usr/local/bin/os-proxy",
      "mode": 493,
      "contents": {
        "source": "${osproxy_url}",
        "verification": {
          "hash": "sha512-${osproxy_sha512}"
        }
      }
    }]
  },
  "systemd": {
    "units": [{
      "name": "os-proxy.service",
      "enabled": true,
      "contents": "[Service]\nType=simple\nUser=osproxy\nWorkingDirectory=/var/home/osproxy\nExecStart=/usr/local/bin/os-proxy -authurl='${auth_url}' -proxyurl='${proxy_url}'\n\n[Install]\nWantedBy=multi-user.target\n"
    }]
  }
}
EOF

sg_id="$(openstack security group create -f value -c id "$name")"
>&2 echo "Created security group '${sg_id}'"
openstack security group rule create --ingress --protocol tcp  --dst-port "$port" --description "${name} tcp in ${port}" "$sg_id" >/dev/null
openstack security group rule create --ingress --protocol icmp --dst-port "$port" --description "${name} ping" "$sg_id" >/dev/null
>&2 echo "Security group rule created for port '${port}'"

network_id="$(openstack network create -f value -c id "$name")"
>&2 echo "Created network '${network_id}'"

subnet_id="$(openstack subnet create -f value -c id \
		--network "$network_id" \
		--subnet-range '172.28.84.0/24' \
		"$name")"
>&2 echo "Created subnet '${subnet_id}'"

router_id="$(openstack router create -f value -c id \
		"$name")"
>&2 echo "Created router '${router_id}'"
openstack router add subnet "$router_id" "$subnet_id"
openstack router set --external-gateway "$external_network" "$router_id"

port_id="$(openstack port create -f value -c id \
		--network "$network_id" \
		--security-group "$sg_id" \
		"$name")"
>&2 echo "Created port '${port_id}'"

server_id="$(openstack server create -f value -c id \
		--image "$server_image" \
		--flavor "$server_flavor" \
		--nic "port-id=$port_id" \
		--security-group "$sg_id" \
		--user-data "$ignition" \
		"$name")"
>&2 echo "Created server '${server_id}'"

fip_id="$(openstack floating ip create -f value -c id \
		--description "$name FIP" \
		"$external_network")"
>&2 echo "Created floating IP '${fip_id}' $(openstack floating ip show -f value -c floating_ip_address "$fip_id")"
openstack server add floating ip "$server_id" "$fip_id"

>&2 echo -n 'Waiting for the proxy to be ready'
declare proxy_ready=no
while [[ $proxy_ready != yes ]]; do
	curl -ikX POST \
		--connect-timeout 5 \
		-H 'Accept: application/json' \
		-H 'Content-Type: application/json' \
		-H 'User-Agent: proxy-deployer' \
		"${proxy_url}/v3/auth/tokens" >/dev/null 2>&1 && proxy_ready=yes ||\
		>&2 echo -n '.'
done
>&2 echo

>&2 echo "Proxy responding on '${proxy_url}'. Press ENTER to tear down."
read pause
