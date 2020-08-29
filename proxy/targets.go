package proxy

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
)

type Target struct {
	Scheme string
	Host   string
}

func NewTarget(u url.URL) Target {
	return Target{
		Scheme: u.Scheme,
		Host:   u.Host,
	}
}

func (t Target) Rewrite(u *url.URL, targetPath string) {
	u.Scheme = t.Scheme
	u.Host = t.Host
	u.Path = targetPath
}

// https://golang.org/src/net/http/httputil/reverseproxy.go#L102
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// https://golang.org/src/net/http/httputil/reverseproxy.go#L114
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()
	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")
	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

type AddressBook struct {
	sync.RWMutex
	targets map[string]Target

	proxy Target
}

// proxy is scheme, host and port of this proxy. All OpenStack URLs in the
// catalog will be rewritten using this base URL.
func NewAddressBook(proxyURL string) (*AddressBook, error) {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &AddressBook{
		targets: make(map[string]Target),
		proxy: Target{
			Scheme: proxy.Scheme,
			Host:   proxy.Host,
		},
	}, nil
}

func splitAlias(path string) (alias, targetPath string) {
	s := strings.SplitN(path, "/", 3)
	if len(s) < 2 {
		return "", "/"
	}
	if len(s) == 2 {
		return s[1], "/"
	}
	return s[1], "/" + s[2]
}

// Resolve rewrites the passed URL with the target found in the map. Errors if
// the target is not in the map.
func (a *AddressBook) Rewrite(reqURL *url.URL) error {
	a.RLock()
	defer a.RUnlock()

	alias, targetPath := splitAlias(reqURL.Path)

	t, ok := a.targets[alias]
	if !ok {
		return fmt.Errorf("unkown target: %q", reqURL)
	}

	t.Rewrite(reqURL, targetPath)

	if alias == "v3" {
		reqURL.Path = "/v3" + reqURL.Path
	}
	if alias == "v2" {
		reqURL.Path = "/v2" + reqURL.Path
	}

	return nil
}

func (a *AddressBook) Set(alias string, target url.URL) {
	a.Lock()
	defer a.Unlock()

	a.targets[alias] = NewTarget(target)
}

func (a *AddressBook) Alias(u *url.URL) error {
	a.Lock()
	defer a.Unlock()

	newT := NewTarget(*u)

	for alias, t := range a.targets {
		if t == newT {
			// The target is known already
			u.Scheme = a.proxy.Scheme
			u.Host = a.proxy.Host
			u.Path = "/" + alias + u.Path
			return nil
		}
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %v", err)
	}
	alias := id.String()

	a.targets[alias] = NewTarget(*u)

	a.proxy.Rewrite(u, "/"+alias+u.Path)

	return nil
}
