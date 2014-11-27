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

	k8sConfig := &k8sclient.Config{
		Host:    options.KubernetesMaster,
		Version: options.KubernetesApiVersion,
	}

	k8sClient, err := k8sclient.New(k8sConfig)
	if err != nil {
		log.Panic(err)
	}

	if serverVersion, err := k8sClient.ServerVersion(); err != nil {
		log.Panic("Couldn't retrieve Kubernetes server version - incorrect URL possibly? ", err)
	} else {
		log.Printf("Connecting to Kubernetes server at version %v", serverVersion.String())
	}

	restful.DefaultContainer.Filter(NCSACommonLogFormatLogger())

	config := swagger.Config{
		WebServices: restful.RegisteredWebServices(),
		ApiPath:     prefix + "/apidocs.json",
	}
	swagger.InstallSwaggerService(config)

	resources := []SelfRegisteringResource{
		PingResource{},
	}
	for _, resource := range resources {
		resource.Register(prefix)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", options.Port), nil))
}
