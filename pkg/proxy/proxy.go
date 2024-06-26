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

package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
)

type contextKey string

const urlKey contextKey = "original url"

var urlRE = regexp.MustCompile(`"url":\s?"(.*?)"`)

type loggingTransport struct {
	transport http.RoundTripper
}

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	originalURL := req.Context().Value(urlKey).(*url.URL)
	reqMethod := req.Method
	reqURL := req.URL
	req.Header.Del("X-Forwarded-For")
	req.Host = req.URL.Host
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		log.Printf("%s %q -> %q FAILED", reqMethod, originalURL, reqURL)
		return nil, err
	}
	log.Printf("%s %q -> %q %s", reqMethod, originalURL, reqURL, res.Status)
	res.Header.Del("strict-transport-security")
	return res, err
}

func NewOpenstackProxyHandler(proxyURL, identityEndpoint string, tlsConfig *tls.Config) (*httputil.ReverseProxy, error) {
	osAuthURL, err := url.Parse(identityEndpoint)
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

	transport := http.DefaultTransport
	if tlsConfig != nil {
		transport.(*http.Transport).TLSClientConfig = tlsConfig
	}

	return &httputil.ReverseProxy{
		Transport: loggingTransport{transport: transport},
		Director: func(req *http.Request) {
			ctx := context.WithValue(req.Context(), urlKey, req.URL)
			*req = *req.Clone(ctx)
			if req.URL.Path == "/" {
				req.URL.Path = "/v3"
			}
			err := addressBook.Rewrite(req.URL)
			if err != nil {
				panic(err)
			}
		},
		ModifyResponse: func(res *http.Response) error {
			if reqURL := res.Request.URL; reqURL.Scheme == osAuthURL.Scheme && reqURL.Host == osAuthURL.Host && reqURL.Path == "/v3/auth/tokens" {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					return err
				}
				newBody := urlRE.ReplaceAllFunc(body, rewriteFunc)
				res.Body = io.NopCloser(bytes.NewReader(newBody))
				res.ContentLength = int64(len(newBody))
				res.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
			}
			return nil
		},
	}, nil
}
