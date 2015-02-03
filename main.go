package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	flags "github.com/jessevdk/go-flags"
)

var logger *log.Logger = log.New(os.Stdout, "", 0)

const prefix = "/api"

type Options struct {
	Port                 uint16 `short:"p" long:"port" description:"The port to listen on" default:"9090"`
	KubernetesMaster     string `short:"k" long:"kubernetes-master" description:"The URL to the Kubernetes master"`
	KubernetesApiVersion string `short:"v" long:"kubernetes-api-version" description:"The version of the Kubernetes API to use" default:"v1beta2"`
	Insecure             bool   `long:"insecure" description:"Trust all server certificates" default:"false"`
	StaticDir            string `short:"w" long:"www" description:"Optional directory to serve static files from"`
	Html5Mode            bool   `long:"html5mode" description:"Send default page (/index.html) on 404" default:"true"`
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

	if err := newApiProxyServer(k8sConfig); err != nil {
		log.Panic("Couldn't start API proxy server", err)
	}

	http.Handle("/proxy/", NewRequestForwarder(k8sClient))

	if len(options.StaticDir) > 0 {
		defaultPage := "/"
		if !options.Html5Mode {
			defaultPage = ""
		}
		http.Handle("/", FileServerWithDefault(http.Dir(options.StaticDir), defaultPage))
	}

	log.Printf("Listening on port %d", options.Port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", options.Port), nil))
}
