build/container: stage/k8s-rest-extras Dockerfile
	docker build --no-cache -t k8s-rest-extras .
	touch build/container

build/k8s-rest-extras: *.go
	GOOS=linux GOARCH=amd64 go build -o build/k8s-rest-extras

stage/k8s-rest-extras: build/k8s-rest-extras
	mkdir -p stage
	cp build/k8s-rest-extras stage/k8s-rest-extras

release:
	docker tag k8s-rest-extras jimmidyson/k8s-rest-extras
	docker push jimmidyson/k8s-rest-extras

.PHONY: clean
clean:
	rm -rf build
