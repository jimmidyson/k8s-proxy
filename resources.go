package main

import (
	"io"

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
