FROM flynn/busybox
MAINTAINER Jimmi Dyson <jimmidyson@gmail.com>

ADD ./build/k8s-proxy /bin/k8s-proxy

EXPOSE 9090

ENTRYPOINT ["/bin/k8s-proxy"]
CMD []
