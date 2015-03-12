FROM gliderlabs/alpine:3.1
MAINTAINER Jimmi Dyson <jimmidyson@gmail.com>
ENTRYPOINT ["/bin/k8s-proxy"]
EXPOSE 9090

COPY . /go/src/github.com/jimmidyson/k8s-proxy

RUN apk-install go git mercurial ca-certificates openssl \
  && export GOPATH=/go \
  && export PATH=${GOPATH}/bin:${PATH} \
  && cd ${GOPATH}/src/github.com/jimmidyson/k8s-proxy \
  && go get github.com/tools/godep \
  && godep get \
  && godep go build -ldflags "-X main.Version $(cat VERSION)" -o /bin/k8s-proxy \
  && rm -rf ${GOPATH} \
  && apk del go git mercurial
