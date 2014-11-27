FROM busybox:latest
MAINTAINER Jimmi Dyson <jimmidyson@gmail.com>

ADD ./stage/k8s-rest-extras /bin/k8s-rest-extras

EXPOSE 8000

ENTRYPOINT ["/bin/k8s-rest-extras"]
CMD []
