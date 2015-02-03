# Kubernetes Proxy

Provides a way to run a proxy to the Kubernetes API server as well as providing a
proxy to running services & pods via their Kubernetes resource names.

You can also set a directory to serve static files from so you can easily create
a web app that uses the proxy endpoints. With the popularity of Angular, you can also
specify to return `/index.html` instead of 404s. This makes it easy to use `html5mode`.

## Running

````
Usage:
  k8s-proxy [OPTIONS]

Application Options:
  -p, --port=                   The port to listen on (9090)
  -k, --kubernetes-master=      The URL to the Kubernetes master
  -v, --kubernetes-api-version= The version of the Kubernetes API to use
                                (v1beta2)
      --insecure                Trust all server certificates (false)
  -w, --www=                    Optional directory to serve static files from
      --html5mode               Send default page (/index.html) on 404 (true)

Help Options:
  -h, --help                    Show this help message
```

A Docker image (`jimmidyson/k8s-proxy`) is also provided to make it easy to layer on your own static content.

## Proxy URLs

`/api/v1beta1/*`, `/api/v1beta2/*` - proxied straight through to the specified Kubernetes API server

`/proxy/<namespace>/service/<serviceId>/<path>` - proxied through to the specified service
`/proxy/<namespace>/pod/<podId>/<port>/<path>` - proxied through to the specified pod on the specified port
