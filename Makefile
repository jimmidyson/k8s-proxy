NAME=k8s-proxy
VERSION=$(shell cat VERSION)

dev:
	@docker history $(NAME):dev &> /dev/null \
		|| docker build -f Dockerfile.dev -t $(NAME):dev .
	@docker run --rm \
		-v $(PWD):/go/src/github.com/jimmidyson/$(NAME) \
		-p 9090:9090 \
		$(NAME):dev

local: *.go
	godep go build -ldflags "-X main.Version $(VERSION)-dev" -o build/k8s-proxy

build:
	mkdir -p build
	docker build -t $(NAME):$(VERSION) .
	docker save $(NAME):$(VERSION) | gzip -9 > build/$(NAME)_$(VERSION).tgz

release:
	rm -rf release && mkdir release
	go get github.com/progrium/gh-release/...
	cp build/* release
	gh-release create jimmidyson/$(NAME) $(VERSION) \
		$(shell git rev-parse --abbrev-ref HEAD) $(VERSION)

.PHONY: dev build release
