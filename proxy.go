package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
)

func newApiProxyServer(cfg *k8sclient.Config) error {
	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "/api"
	}
	target, err := url.Parse(singleJoiningSlash(cfg.Host, prefix))
	if err != nil {
		return err
	}
	proxy := newProxyServer(target)
	if proxy.Transport, err = k8sclient.TransportFor(cfg); err != nil {
		return err
	}
	http.Handle("/api/v1beta1/", http.StripPrefix("/api/", proxy))
	http.Handle("/api/v1beta2/", http.StripPrefix("/api/", proxy))
	return nil
}

func newProxyServer(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	}
	return &httputil.ReverseProxy{Director: director}
}

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
