package main

import (
	"io"
	"net/http"

	k8sapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	restful "github.com/emicklei/go-restful"
)

type SelfRegisteringResource interface {
	Register(string)
}

type PingResource struct {
}

func (r PingResource) Register(prefix string) {
	ws := new(restful.WebService)
	ws.Path(prefix + "/ping")
	ws.Route(ws.GET("/").To(pong))
	restful.Add(ws)
}

func pong(req *restful.Request, resp *restful.Response) {
	io.WriteString(resp, "pong")
}

type ContainerResource struct {
	client *k8sclient.Client
}

func (r ContainerResource) Register(prefix string) {
	ws := new(restful.WebService)
	ws.Path(prefix + "/containers").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/").To(r.findContainer).Writes(k8sapi.MinionList{}))
	restful.Add(ws)
}

func (r ContainerResource) findContainer(request *restful.Request, response *restful.Response) {
	if minions, err := r.client.Minions().List(); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteEntity(minions.Items)
	}
}
