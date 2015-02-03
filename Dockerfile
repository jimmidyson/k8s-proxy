FROM busybox:latest
MAINTAINER Jimmi Dyson <jimmidyson@gmail.com>

ADD ./stage/k8s-proxy /bin/k8s-proxy

EXPOSE 9090

ENTRYPOINT ["/bin/k8s-proxy"]
CMD []
