build/k8s-proxy: *.go
	godep go build -o build/k8s-proxy

cross:
	GOOS=linux GOARCH=amd64 godep go build -o build/k8s-proxy-linux64
	GOOS=darwin GOARCH=amd64 godep go build -o build/k8s-proxy-darwin64
	GOOS=freebsd GOARCH=amd64 godep go build -o build/k8s-proxy-darwin64
	GOOS=windows GOARCH=amd64 godep go build -o build/k8s-proxy-windows64.exe

image:
	docker build --no-cache -t k8s-proxy .

release:
	docker tag k8s-proxy jimmidyson/k8s-proxy
	docker push jimmidyson/k8s-proxy

.PHONY: clean
clean:
	rm -rf build
