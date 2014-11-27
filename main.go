package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	flags "github.com/jessevdk/go-flags"
)

var logger *log.Logger = log.New(os.Stdout, "", 0)

const prefix = "/api"

type Options struct {
	Port                 uint16 `short:"p" long:"port" description:"The port to listen on" default:"9090"`
	KubernetesMaster     string `short:"k" long:"kubernetes-master" description:"The URL to the Kubernetes master"`
	KubernetesApiVersion string `short:"v" long:"kubernetes-api-version" description:"The version of the Kubernetes API to use" default:"v1beta2"`
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

	if len(options.KubernetesMaster) == 0 && len(os.Getenv("KUBERNETES_RO_SERVICE_HOST")) > 0 {
		options.KubernetesMaster = os.ExpandEnv("http://${KUBERNETES_RO_SERVICE_HOST}:${KUBERNETES_RO_SERVICE_PORT}")
	}

	k8sConfig := &k8sclient.Config{
		Host:    options.KubernetesMaster,
		Version: options.KubernetesApiVersion,
	}

	k8sClient, err := k8sclient.New(k8sConfig)
	if err != nil {
		log.Panic(err)
	}

	if serverVersion, err := k8sClient.ServerVersion(); err != nil {
		log.Panic("Couldn't retrieve Kubernetes server version - incorrect URL possibly?", err)
	} else {
		log.Printf("Connecting to Kubernetes master at %v running version %v", options.KubernetesMaster, serverVersion.String())
	}

	config := swagger.Config{
		WebServices: restful.RegisteredWebServices(),
		ApiPath:     prefix + "/apidocs.json",
	}
	swagger.InstallSwaggerService(config)

	resources := []SelfRegisteringResource{
		PingResource{},
		ContainerResource{client: k8sClient},
	}
	for _, resource := range resources {
		resource.Register(prefix)
	}

	restful.Filter(NCSACommonLogFormatLogger())
	restful.Filter(restful.OPTIONSFilter())
	restful.DefaultContainer.EnableContentEncoding(true)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", options.Port), nil))
}
