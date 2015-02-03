package main

import (
	"io"
	"net/http"

	k8sapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	k8sclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
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
	ws.Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	ws.Path(prefix)
	ws.Route(ws.GET("/minions").To(r.minions).Writes(k8sapi.NodeList{}))
	ws.Route(ws.GET("/pods").To(r.pods).Writes(k8sapi.PodList{}))
	ws.Route(ws.GET("/{namespace}/pods").To(r.pods).Writes(k8sapi.PodList{}))
	ws.Route(ws.GET("/services").To(r.services).Writes(k8sapi.ServiceList{}))
	ws.Route(ws.GET("/{namespace}/services").To(r.services).Writes(k8sapi.ServiceList{}))
	ws.Route(ws.GET("/replicationControllers").To(r.replicationControllers).Writes(k8sapi.ReplicationControllerList{}))
	ws.Route(ws.GET("/{namespace}/replicationControllers").To(r.replicationControllers).Writes(k8sapi.ReplicationControllerList{}))
	restful.Add(ws)
}

func (r ContainerResource) minions(request *restful.Request, response *restful.Response) {
	if minions, err := r.client.Nodes().List(); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteEntity(minions.Items)
	}
}

func (r ContainerResource) pods(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	if pods, err := r.client.Pods(namespace).List(labels.Everything()); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteEntity(pods.Items)
	}
}

func (r ContainerResource) services(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	if services, err := r.client.Services(namespace).List(labels.Everything()); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteEntity(services.Items)
	}
}

func (r ContainerResource) replicationControllers(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	if rcs, err := r.client.ReplicationControllers(namespace).List(labels.Everything()); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	} else {
		response.WriteEntity(rcs.Items)
	}
}
