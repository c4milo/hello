BUILD_DIR ?= public
BIN_DIR ?= node_modules/.bin
BINDATA_TARGET = embedded.go
ENV ?= production

dev: ## Starts webpack dev server which also proxies requests to API server.
	npm start

clean: ## Cleans up generated files
	rm -rf $(BUILD_DIR)/{app,vendor}.{js,css}

dist: clean ## Generates a distributable artifact for front-end code
	NODE_ENV=$(ENV) $(BIN_DIR)/webpack --progress

serve: dist ## Starts up a static server only, in case this project is deployed separately from the API.
	$(BIN_DIR)/static $(BUILD_DIR) --spa

embed: dist ## Embeds static assets within the Go binary, it could depend on the dist target too in order to build the SPA app.
	go-bindata -prefix "public" -o $(BINDATA_TARGET) -pkg static public/...
	gofmt -s -w $(BINDATA_TARGET)
	goimports -w $(BINDATA_TARGET)

.PHONY: dist dev clean serve embed
