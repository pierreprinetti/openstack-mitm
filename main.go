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
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/pierreprinetti/openstack-mitm/proxy"
)

const (
	defaultPort = "5443"
)

var (
	proxyURLstring = flag.String("proxy-url", "", "The address this proxy will be reachable at")
	osAuthURL      = flag.String("remote-authurl", "", "OpenStack entrypoint (OS_AUTH_URL)")
	osCaCert       = flag.String("remote-cacert", "", "OpenStack CA certificate (OS_CACERT)")
	insecure       = flag.Bool("insecure", false, "Insecure connection to OpenStack")
)

func init() {
	flag.Parse()

	var errexit bool
	if *proxyURLstring == "" {
		errexit = true
		log.Print("Missing required parameter: --proxyurl")
	}

	if *osAuthURL == "" {
		log.Print("Missing required parameter: --authurl")
	}

	if errexit {
		log.Fatal("Exiting.")
	}
}

func main() {
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

	transport := http.DefaultTransport.(*http.Transport)

	if caCertPath := *osCaCert; caCertPath != "" {
		b, err := os.ReadFile(caCertPath)
		if err != nil {
			log.Fatalf("Failed to read the given PEM certificate: %v", err)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(b) {
			log.Fatal("Failed to parse the given PEM certificate")
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	}

	if *insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: *insecure}
	}

	p, err := proxy.NewOpenstackProxy(proxyURL.String(), *osAuthURL, transport)
	if err != nil {
		panic(err)
	}

	{
		if err := generateCertificate(proxyURL.Hostname()); err != nil {
			log.Fatal(err)
		}
		log.Printf("Certificate correctly generated for %q", proxyURL.Hostname())
	}

	log.Printf("Proxying to %q", *osAuthURL)
	log.Printf("Listening on %q", proxyURL)

	log.Fatal(
		http.ListenAndServeTLS(":"+proxyURL.Port(), "cert.pem", "key.pem", p),
	)
}
