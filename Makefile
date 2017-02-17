NAME := hello
VERSION := v1.0.0
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.AppName=$(NAME)"
BLDTAGS := -tags ""
DEVTAGS := -tags "dev"

PROTOC = /usr/local/bin/protoc

help: ## Shows this help text
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

generate: ## Builds and embeds web app in Go binary.
	go generate static/static.go

protoc: ## Generates gRPC code
	$(PROTOC) -I. \
	-I$(GOPATH)/src \
	-I$(GOPATH)/src/github.com \
	-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--grpc-gateway_out=logtostderr=true:. \
	--swagger_out=logtostderr=true:static/public/lib \
	--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. $(NAME).proto

testrace: migrations ## Runs all tests with race detector
	go test $(DEVTAGS) -parallel 2 -race ./...

dev: protoc generate ## Builds a dev binary for local testing.
	go build $(DEVTAGS) $(LDFLAGS) -o $(NAME) cmd/$(NAME)/$(NAME).go

clean: ## Cleans object files
	go clean $(DEVTAGS) $(LDFLAGS)

build: protoc ## Generates a build for linux and darwin into build/ folder
	@rm -rf build/
	@gox $(BLDTAGS) $(LDFLAGS) \
	-os="darwin" \
	-os="linux" \
	-output "build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)" \
	./...

install: protoc ## Installs binary in Go's binary folder
	go install $(DEVTAGS) $(LDFLAGS)

dist: build ## Generates distributable artifacts in dist/ folder
	$(eval FILES := $(shell ls build))
	@rm -rf dist && mkdir dist
	@for f in $(FILES); do \
		(cd $(shell pwd)/build/$$f && tar -cvzf ../../dist/$$f.tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $$f.tar.gz > $$f.sha512); \
		echo $$f; \
	done

release: test dist ## Generates a release in Github and uploads artifacts.
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges --reverse); \
	github-release hooklift/$(NAME) $(VERSION) "$$(git rev-parse --abbrev-ref HEAD)" "**Changelog**<br/>$$changelog" 'dist/*'; \
	git pull

devcerts: ## Generates dev TLS certificates
	openssl ecparam -genkey -name secp384r1 -out certs/server-key.pem && \
	openssl req -new -x509 -key certs/server-key.pem -out certs/server.pem -days 36000


.PHONY: help dev build protoc install deps dist release

.DEFAULT_GOAL := help
