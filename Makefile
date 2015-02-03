build/container: stage/k8s-proxy Dockerfile
	docker build --no-cache -t k8s-proxy .
	touch build/container

build/k8s-proxy: *.go
	GOOS=linux GOARCH=amd64 godep go build -o build/k8s-proxy

stage/k8s-proxy: build/k8s-proxy
	mkdir -p stage
	cp build/k8s-proxy stage/k8s-proxy

release:
	docker tag k8s-proxy jimmidyson/k8s-proxy
	docker push jimmidyson/k8s-proxy

.PHONY: clean
clean:
	rm -rf build
