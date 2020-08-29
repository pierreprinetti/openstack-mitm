package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/BurntSushi/xdg"
	"github.com/shiftstack/os-proxy/proxy"
	"gopkg.in/yaml.v2"
)

const (
	defaultPort = "5443"
)

var (
	proxyURL  = flag.String("proxyurl", "", "The host this proxy will be reachable")
	osAuthURL = flag.String("authurl", "", "OpenStack entrypoint (OS_AUTH_URL)")
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

	if *proxyURL == "" {
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

	var proxyHost, proxyPort string
	{
		u, err := url.Parse(*proxyURL)
		if err != nil {
			log.Fatal(err)
		}
		proxyHost = u.Host
		proxyPort = u.Port()
		if proxyPort == "" {
			proxyPort = defaultPort
		}
	}

	p, err := proxy.NewOpenstackProxy(*proxyURL, *osAuthURL)
	if err != nil {
		panic(err)
	}

	log.Printf("Rewriting URLs to %q", proxyHost)
	log.Printf("Proxying %q", *osAuthURL)

	{
		proxyDomain := strings.SplitN(proxyHost, ":", 2)[0]
		if err := generateCertificate(proxyDomain); err != nil {
			log.Fatal(err)
		}
		log.Printf("Certificate correctly generated for %q", proxyDomain)
	}

	log.Printf("Starting the server on %q...", proxyPort)
	log.Fatal(
		http.ListenAndServeTLS(":"+proxyPort, "cert.pem", "key.pem", p),
	)
}
