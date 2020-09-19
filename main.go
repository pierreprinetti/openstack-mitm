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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/BurntSushi/xdg"
	"github.com/shiftstack/os-proxy/proxy"
	"gopkg.in/yaml.v2"
)

const (
	defaultPort = "5443"
)

var (
	proxyURLstring = flag.String("proxyurl", "", "The host this proxy will be reachable")
	osAuthURL      = flag.String("authurl", "", "OpenStack entrypoint (OS_AUTH_URL)")
)

func getAuthURL(cloudName string) (string, error) {
	cloudsPath, err := xdg.Paths{XDGSuffix: "openstack"}.ConfigFile("clouds.yaml")
	if err != nil {
		return "", err
	}

	cloudsFile, err := os.Open(cloudsPath)
	if err != nil {
		return "", err
	}
	defer cloudsFile.Close()

	var clouds struct {
		Clouds map[string]struct {
			Auth struct {
				URL string `yaml:"auth_url"`
			} `yaml:"auth"`
		} `yaml:"clouds"`
	}

	cloudsContent, err := ioutil.ReadAll(cloudsFile)
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(cloudsContent, &clouds); err != nil {
		return "", err

	}

	if _, ok := clouds.Clouds[cloudName]; !ok {
		return "", fmt.Errorf("cloud %q not found", cloudName)
	}

	return clouds.Clouds[cloudName].Auth.URL, nil
}

func main() {
	flag.Parse()

	if *proxyURLstring == "" {
		log.Fatal("Missing required --proxyurl parameter")
	}

	if *osAuthURL == "" {
		if osCloud := os.Getenv("OS_CLOUD"); osCloud != "" {
			log.Print("Missing --authurl parameter; parsing from clouds.yaml")
			var err error
			*osAuthURL, err = getAuthURL(osCloud)
			if err != nil {
				log.Fatalf("Error parsing auth_url from cloud %q: %v", osCloud, err)
			}
		} else {
			log.Fatal("Missing --authurl parameter. Set it or set the relative OS_CLOUD environment to enable parsing it from clouds.yaml.")
		}
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
