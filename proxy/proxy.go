package proxy

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
)

const debug = false

var urlRE = regexp.MustCompile(`"url":\s?"(.*?)"`)

func NewOpenstackProxy(proxyURL, osAuth string) (*httputil.ReverseProxy, error) {
	osAuthURL, err := url.Parse(osAuth)
	if err != nil {
		return nil, err
	}
	addressBook, err := NewAddressBook(proxyURL)
	if err != nil {
		return nil, err
	}
	addressBook.Set("v3", *osAuthURL)
	addressBook.Set("v2", *osAuthURL)

	rewriteFunc := func(src []byte) []byte {
		found := urlRE.FindSubmatch(src)
		u, err := url.Parse(string(found[1]))
		if err != nil {
			panic(err)
		}
		if err := addressBook.Alias(u); err != nil {
			panic(err)
		}
		return []byte(`"url": "` + u.String() + `"`)
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			if debug {
				log.Printf("req in: %s\n", req.URL)
			}
			if req.URL.Path == "/" {
				req.URL.Path = "/v3"
			}
			err := addressBook.Rewrite(req.URL)
			if err != nil {
				panic(err)
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
		ModifyResponse: func(res *http.Response) error {
			if debug {
				log.Printf("req out: %s\n", res.Request.URL)
			}
			if reqURL := res.Request.URL; reqURL.Scheme == osAuthURL.Scheme && reqURL.Host == osAuthURL.Host && reqURL.Path == "/v3/auth/tokens" {
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					return err
				}
				newBody := urlRE.ReplaceAllFunc(body, rewriteFunc)
				res.Body = ioutil.NopCloser(bytes.NewReader(newBody))
				res.ContentLength = int64(len(newBody))
				res.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
			}
			return nil
		},
	}, nil
}
