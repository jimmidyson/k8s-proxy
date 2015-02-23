# Kubernetes Proxy [![Circle CI](https://circleci.com/gh/jimmidyson/k8s-proxy.svg?style=svg)](https://circleci.com/gh/jimmidyson/k8s-proxy)

Provides a way to run a proxy to the Kubernetes API server as well as providing a
proxy to running services & pods via their Kubernetes resource names.

You can also set a directory to serve static files from so you can easily create
a web app that uses the proxy endpoints. With the popularity of Angular, you can also
specify to return `/index.html` instead of 404s. This makes it easy to use `html5mode`.

Finally, specify `tls-cert` & `tls-key` to listen over TLS with http2 enabled.

## Running

```
Usage:
  k8s-proxy [OPTIONS]

Application Options:
  -p, --port=                   The port to listen on (9090)
  -k, --kubernetes-master=      The URL to the Kubernetes master
  -v, --kubernetes-api-version= The version of the Kubernetes API to use (v1beta2)
      --insecure                Trust all server certificates (false)
  -w, --www=                    Optional directory to serve static files from (.)
      --www-prefix=             Prefix to serve static files on (/)
      --api-prefix=             Prefix to serve Kubernetes API on (/api/)
      --404=                    Page to send on 404 (useful for e.g. Angular html5mode
                                default page)
      --tls-cert=               TLS cert file
      --tls-key=                TLS key file

Help Options:
  -h, --help                    Show this help message
```

A Docker image (`jimmidyson/k8s-proxy`) is also provided to make it easy to layer on your own static content.

## Kubernetes API URLs

`/api/v1beta1/*`, `/api/v1beta2/*`, `/api/v1beta3/*` URLs are proxied straight through to the specified Kubernetes API server.

Note that the prefix `/api` can be changed by using the `--api-prefix` flag.

### Pod & service proxying

The Kubernetes API exposes URLs that will proxy requests to services & pods. The format for these is slightly different
between `v1beta3` & previous versions due to namespace changes.

To proxy to a service use any the following URLs:

`/api/v1beta1/proxy/services/<serviceId>/<path>?namespace=<namespace>`
`/api/v1beta2/proxy/services/<serviceId>/<path>?namespace=<namespace>`
`/api/v1beta3/proxy/ns/<namespace>/services/<serviceId>/<path>`

To proxy to a pod use any of the following URLs:

`/api/v1beta1/proxy/pods/<podId>:<port>/<path>?namespace=<namespace>`
`/api/v1beta2/proxy/pods/<podId>:<port>/<path>?namespace=<namespace>`
`/api/v1beta3/proxy/ns/<namespace>/pods/<podId>:<port>/<path>`
