package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"

	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	"github.com/bradfitz/http2"
	flags "github.com/jessevdk/go-flags"
)

var logger *log.Logger = log.New(os.Stdout, "", 0)

const prefix = "/api"

type Options struct {
	Port                 uint16 `short:"p" long:"port" description:"The port to listen on" default:"9090"`
	KubernetesMaster     string `short:"k" long:"kubernetes-master" description:"The URL to the Kubernetes master"`
	KubernetesApiVersion string `short:"v" long:"kubernetes-api-version" description:"The version of the Kubernetes API to use" default:"v1beta2"`
	Insecure             bool   `long:"insecure" description:"Trust all server certificates" default:"false"`
	StaticDir            string `short:"w" long:"www" description:"Optional directory to serve static files from" default:"."`
	StaticPrefix         string `long:"www-prefix" description:"Prefix to serve static files on" default:"/"`
	ApiPrefix            string `long:"api-prefix" description:"Prefix to serve Kubernetes API on" default:"/api/"`
	Error404             string `long:"404" description:"Page to send on 404 (useful for e.g. Angular html5mode default page)"`
}

func main() {
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	if _, err := parser.Parse(); err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(options.KubernetesMaster) == 0 && len(os.Getenv("KUBERNETES_SERVICE_HOST")) > 0 {
		options.KubernetesMaster = "https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}"
	}

	options.KubernetesMaster = os.ExpandEnv(options.KubernetesMaster)

	k8sConfig := &k8sclient.Config{
		Host:     options.KubernetesMaster,
		Version:  options.KubernetesApiVersion,
		Insecure: options.Insecure,
	}

	k8sClient, err := k8sclient.New(k8sConfig)
	if err != nil {
		log.Panic(err)
	}

	if serverVersion, err := k8sClient.ServerVersion(); err != nil {
		log.Panic("Couldn't retrieve Kubernetes server version - incorrect URL?", err)
	} else {
		log.Printf("Connecting to Kubernetes master at %v running version %v", options.KubernetesMaster, serverVersion.String())
	}

	// Add SVG mimetype...
	mime.AddExtensionType(".svg", "image/svg+xml")

	_, err = kubectl.NewProxyServer(options.StaticDir, options.ApiPrefix, options.StaticPrefix, k8sConfig)

	log.Printf("Listening on port %d", options.Port)

	http2.VerboseLogs = true

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", options.Port),
	}

	http2.ConfigureServer(srv, &http2.Server{})

	if len(options.Error404) > 0 {
		srv.Handler = Handle404(http.DefaultServeMux, http.Dir(options.StaticDir), options.Error404)
	}
	srv.ListenAndServeTLS("mycert1.cer", "mycert1.key")
}

type hijack404 struct {
	http.ResponseWriter
	r            *http.Request
	fs           http.FileSystem
	error404Page string
	handled      bool
}

func (h *hijack404) Write(p []byte) (int, error) {
	if h.handled {
		f, err := h.fs.Open(h.error404Page)
		if err != nil {
			h.ResponseWriter.Write([]byte("404 page not found"))
			return 0, errors.New("404 page not found")
		}
		_, err = f.Stat()
		if err != nil {
			h.ResponseWriter.Write([]byte("404 page not found"))
			return 0, errors.New("404 page not found")
		}
		contents, err := ioutil.ReadAll(f)
		ctype := http.DetectContentType(contents)
		h.ResponseWriter.Header().Set("Content-Type", ctype)
		h.ResponseWriter.Write(contents)
		return 0, nil
	}
	return h.ResponseWriter.Write(p)
}

func (h *hijack404) WriteHeader(code int) {
	if code == http.StatusNotFound {
		h.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
		h.handled = true
	}
	h.ResponseWriter.WriteHeader(code)
}

// Handle404 will pass any 404's from the handler to the handle404
// function. If handle404 returns true, the response is considered complete,
// and the processing by handler is aborted.
func Handle404(handler http.Handler, fs http.FileSystem, error404Page string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijack := &hijack404{ResponseWriter: w, r: r, fs: fs, error404Page: error404Page}
		handler.ServeHTTP(hijack, r)
	})
}
