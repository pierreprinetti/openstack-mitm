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
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud/v2/openstack/config/clouds"

	"github.com/pierreprinetti/openstack-mitm/pkg/cloudout"
	"github.com/pierreprinetti/openstack-mitm/pkg/proxy"
)

var (
	mitmURL          *url.URL
	identityEndpoint string
	tlsConfig        *tls.Config

	tlsCertPath string
	tlsKeyPath  string
)

func init() {
	var (
		mitmURLstring   string
		outputCloudPath string
	)

	flag.StringVar(&mitmURLstring, "url", "http://localhost:13000", "The address this MITM proxy will be reachable at")
	flag.StringVar(&tlsCertPath, "cert", "", "Path to the PEM-encoded TLS certificate")
	flag.StringVar(&tlsKeyPath, "key", "", "Path to the PEM-encoded TLS certificate private key")
	flag.StringVar(&outputCloudPath, "o", "", "Path of the clouds.yaml file that points to this MITM proxy (optional)")
	flag.Parse()

	var err error
	mitmURL, err = url.Parse(mitmURLstring)
	if err != nil {
		log.Fatalf("Failed to parse the URL (%q): %v", mitmURLstring, err)
	}

	authOptions, endpointOptions, parsedTLSConfig, err := clouds.Parse()
	if err != nil {
		log.Fatalf("Failed to parse clouds.yaml: %v", err)
	}

	identityEndpoint = authOptions.IdentityEndpoint
	tlsConfig = parsedTLSConfig

	if outputCloudPath != "" {
		f, err := os.Create(outputCloudPath)
		if err != nil {
			log.Fatalf("Failed to create clouds.yaml at the given destination (%q):%v", outputCloudPath, err)
		}
		if err := cloudout.Write(authOptions, endpointOptions, tlsConfig, mitmURL.String(), tlsCertPath, f); err != nil {
			log.Fatalf("Failed to encode the output clouds.yaml: %v", err)
		}
		if err := f.Close(); err != nil {
			log.Fatalf("Failed to finalize the clouds.yaml file: %v", err)
		}
		log.Printf("clouds.yaml written to %q", outputCloudPath)
	}
}

func main() {
	p, err := proxy.NewOpenstackProxyHandler(mitmURL.String(), identityEndpoint, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to build the OpenStack MITM proxy: %v", err)
	}

	listenURL := ":" + mitmURL.Port()

	log.Printf("Proxying to %q", identityEndpoint)
	log.Printf("Listening on %q", listenURL)

	switch strings.ToLower(mitmURL.Scheme) {
	case "http":
		log.Fatal(http.ListenAndServe(listenURL, p))
	case "https":
		log.Fatal(http.ListenAndServeTLS(listenURL, tlsCertPath, tlsKeyPath, p))
	default:
		log.Fatalf("Unknown scheme %q", mitmURL.Scheme)
	}
}
