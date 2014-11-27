package main

import (
	"log"
	"net/http"
	"os"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
)

var logger *log.Logger = log.New(os.Stdout, "", 0)

const prefix = "/api"

func main() {

	restful.DefaultContainer.Filter(NCSACommonLogFormatLogger())

	PingResource{}.Register(prefix)

	config := swagger.Config{
		WebServices: restful.RegisteredWebServices(),
		ApiPath:     prefix + "/apidocs.json",
	}
	swagger.InstallSwaggerService(config)

	log.Fatal(http.ListenAndServe(":9090", nil))
}
