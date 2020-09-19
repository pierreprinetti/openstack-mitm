// Copyright 2020 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"

	"github.com/shiftstack/os-proxy/proxy"
)

const (
	defaultPort = "5443"
)

var (
	proxyURLstring = flag.String("proxyurl", "", "The host this proxy will be reachable")
	osAuthURL      = flag.String("authurl", "", "OpenStack entrypoint (OS_AUTH_URL)")
)

func main() {
	flag.Parse()

	if *proxyURLstring == "" {
		log.Fatal("Missing required --proxyurl parameter.")
	}

	if *osAuthURL == "" {
		log.Fatal("Missing required --authurl parameter.")
	}

	var proxyURL *url.URL
	{
		var err error
		proxyURL, err = url.Parse(*proxyURLstring)
		if err != nil {
			log.Fatal(err)
		}

		if proxyURL.Host == "" {
			log.Fatal("The --proxyurl parameter is invalid. It should be in the form: 'https://host[:port]'.")
		}

		if proxyURL.Path != "" {
			log.Fatal("The --proxyurl URL should have empty path.")
		}

		if proxyURL.Port() == "" {
			proxyURL.Host = proxyURL.Hostname() + ":" + defaultPort
		}
	}

	p, err := proxy.NewOpenstackProxy(proxyURL.String(), *osAuthURL)
	if err != nil {
		panic(err)
	}

	log.Printf("Rewriting URLs to %q", proxyURL)
	log.Printf("Proxying %q", *osAuthURL)

	{
		if err := generateCertificate(proxyURL.Hostname()); err != nil {
			log.Fatal(err)
		}
		log.Printf("Certificate correctly generated for %q", proxyURL.Hostname())
	}

	log.Printf("Starting the server on port %s...", proxyURL.Port())
	log.Fatal(
		http.ListenAndServeTLS(":"+proxyURL.Port(), "cert.pem", "key.pem", p),
	)
}
