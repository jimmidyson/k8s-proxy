package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/golang/glog"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/utils"
	"golang.org/x/net/html"
)

type RequestForwarder struct {
	prefix     string
	client     *k8sclient.Client
	errHandler utils.ErrorHandler
	next       http.Handler
}

func (r *RequestForwarder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	destUrl, err := r.getForwardUrl(req)
	if err != nil {
		r.errHandler.ServeHTTP(w, req, err)
		return
	}

	if strings.HasSuffix(req.URL.Path, "/") {
		destUrl.Path = destUrl.Path + "/"
	}

	newReq, err := http.NewRequest(req.Method, destUrl.String(), req.Body)

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: destUrl.Host})
	proxy.Transport = &proxyTransport{
		proxyScheme: req.URL.Scheme,
		proxyHost:   req.URL.Host,
	}
	proxy.FlushInterval = 200 * time.Millisecond
	proxy.ServeHTTP(w, newReq)
}

func (r *RequestForwarder) getForwardUrl(req *http.Request) (*url.URL, error) {
	parts := r.splitPath(req.URL.Path)

	if parts[0] != "proxy" || len(parts) < 4 {
		return nil, fmt.Errorf("Unable to determine kind and namespace from url, %v", req.URL)
	}

	namespace := parts[1]
	kind := parts[2]
	resourceName := parts[3]
	parts = parts[4:]

	if strings.ToLower(kind) == "pod" {
		port := parts[0]
		parts = parts[1:]
		podClient := r.client.Pods(namespace)
		pod, err := podClient.Get(resourceName)
		if err != nil {
			return nil, err
		}
		return url.Parse("http://" + pod.Status.PodIP + ":" + port + "/" + strings.Join(parts, "/"))
	} else {
		serviceClient := r.client.Services(namespace)
		service, err := serviceClient.Get(resourceName)
		if err != nil {
			return nil, err
		}
		return url.Parse("http://" + service.Spec.PortalIP + ":" + strconv.Itoa(service.Spec.Port) + "/" + strings.Join(parts, "/"))
	}
}

func (r *RequestForwarder) splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func NewRequestForwarder(client *k8sclient.Client) *RequestForwarder {
	fwd, _ := forward.New()

	sf := &RequestForwarder{
		prefix:     "proxy",
		client:     client,
		errHandler: utils.DefaultHandler,
		next:       fwd,
	}

	return sf
}

type proxyTransport struct {
	proxyScheme      string
	proxyHost        string
	proxyPathPrepend string
}

func (t *proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add reverse proxy headers.
	req.Header.Set("X-Forwarded-Uri", t.proxyPathPrepend+req.URL.Path)
	req.Header.Set("X-Forwarded-Host", t.proxyHost)
	req.Header.Set("X-Forwarded-Proto", t.proxyScheme)

	resp, err := http.DefaultTransport.RoundTrip(req)

	if err != nil {
		message := fmt.Sprintf("Error: '%s'\nTrying to reach: '%v'", err.Error(), req.URL.String())
		resp = &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       ioutil.NopCloser(strings.NewReader(message)),
		}
		return resp, nil
	}

	cType := resp.Header.Get("Content-Type")
	cType = strings.TrimSpace(strings.SplitN(cType, ";", 2)[0])
	if cType != "text/html" {
		// Do nothing, simply pass through
		return resp, nil
	}

	return t.fixLinks(req, resp)
}

// updateURLs checks and updates any of n's attributes that are listed in tagsToAttrs.
// Any URLs found are, if they're relative, updated with the necessary changes to make
// a visit to that URL also go through the proxy.
// sourceURL is the URL of the page which we're currently on; it's required to make
// relative links work.
func (t *proxyTransport) updateURLs(n *html.Node, sourceURL *url.URL) {
	if n.Type != html.ElementNode {
		return
	}
	attrs, ok := tagsToAttrs[n.Data]
	if !ok {
		return
	}
	for i, attr := range n.Attr {
		if !attrs.Has(attr.Key) {
			continue
		}
		url, err := url.Parse(attr.Val)
		if err != nil {
			continue
		}

		// Is this URL referring to the same host as sourceURL?
		if url.Host == "" || url.Host == sourceURL.Host {
			url.Scheme = t.proxyScheme
			url.Host = t.proxyHost
			origPath := url.Path

			if strings.HasPrefix(url.Path, "/") {
				// The path is rooted at the host. Just add proxy prepend.
				url.Path = path.Join(t.proxyPathPrepend, url.Path)
			} else {
				// The path is relative to sourceURL.
				url.Path = path.Join(t.proxyPathPrepend, path.Dir(sourceURL.Path), url.Path)
			}

			if strings.HasSuffix(origPath, "/") {
				// Add back the trailing slash, which was stripped by path.Join().
				url.Path += "/"
			}

			n.Attr[i].Val = url.String()
		}
	}
}

// scan recursively calls f for every n and every subnode of n.
func (t *proxyTransport) scan(n *html.Node, f func(*html.Node)) {
	f(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.scan(c, f)
	}
}

// fixLinks modifies links in an HTML file such that they will be redirected through the proxy if needed.
func (t *proxyTransport) fixLinks(req *http.Request, resp *http.Response) (*http.Response, error) {
	origBody := resp.Body
	defer origBody.Close()

	newContent := &bytes.Buffer{}
	var reader io.Reader = origBody
	var writer io.Writer = newContent
	encoding := resp.Header.Get("Content-Encoding")
	switch encoding {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("errorf making gzip reader: %v", err)
		}
		gzw := gzip.NewWriter(writer)
		defer gzw.Close()
		writer = gzw
	// TODO: support flate, other encodings.
	case "":
		// This is fine
	default:
		// Some encoding we don't understand-- don't try to parse this
		glog.Errorf("Proxy encountered encoding %v for text/html; can't understand this so not fixing links.", encoding)
		return resp, nil
	}

	doc, err := html.Parse(reader)
	if err != nil {
		glog.Errorf("Parse failed: %v", err)
		return resp, err
	}

	t.scan(doc, func(n *html.Node) { t.updateURLs(n, req.URL) })
	if err := html.Render(writer, doc); err != nil {
		glog.Errorf("Failed to render: %v", err)
	}

	resp.Body = ioutil.NopCloser(newContent)
	// Update header node with new content-length
	// TODO: Remove any hash/signature headers here?
	resp.Header.Del("Content-Length")
	resp.ContentLength = int64(newContent.Len())

	return resp, err
}

// tagsToAttrs states which attributes of which tags require URL substitution.
// Sources: http://www.w3.org/TR/REC-html40/index/attributes.html
//          http://www.w3.org/html/wg/drafts/html/master/index.html#attributes-1
var tagsToAttrs = map[string]util.StringSet{
	"a":          util.NewStringSet("href"),
	"applet":     util.NewStringSet("codebase"),
	"area":       util.NewStringSet("href"),
	"audio":      util.NewStringSet("src"),
	"base":       util.NewStringSet("href"),
	"blockquote": util.NewStringSet("cite"),
	"body":       util.NewStringSet("background"),
	"button":     util.NewStringSet("formaction"),
	"command":    util.NewStringSet("icon"),
	"del":        util.NewStringSet("cite"),
	"embed":      util.NewStringSet("src"),
	"form":       util.NewStringSet("action"),
	"frame":      util.NewStringSet("longdesc", "src"),
	"head":       util.NewStringSet("profile"),
	"html":       util.NewStringSet("manifest"),
	"iframe":     util.NewStringSet("longdesc", "src"),
	"img":        util.NewStringSet("longdesc", "src", "usemap"),
	"input":      util.NewStringSet("src", "usemap", "formaction"),
	"ins":        util.NewStringSet("cite"),
	"link":       util.NewStringSet("href"),
	"object":     util.NewStringSet("classid", "codebase", "data", "usemap"),
	"q":          util.NewStringSet("cite"),
	"script":     util.NewStringSet("src"),
	"source":     util.NewStringSet("src"),
	"video":      util.NewStringSet("poster", "src"),

	// TODO: css URLs hidden in style elements.
}
