package main

import (
	"log"
	"net/http"
	"os"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	flags "github.com/jessevdk/go-flags"
)

var logger *log.Logger = log.New(os.Stdout, "", 0)

const prefix = "/api"

type Options struct {
	KubernetesMaster string `short:"k" long:"kubernetes-master" description:"The URL to the Kubernetes master" default:"http://localhost:8080"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}
		os.Exit(0)
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}
