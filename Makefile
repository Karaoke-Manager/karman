OPENAPI_FILES := $(shell find api -type f -name '*.yaml')
OPENAPI_SPECS := api/karman.yaml $(shell find api/tags -type f -name '*.yaml')
GO_FILES = $(shell find . -type f -name '*.go')


.PHONY: all
all: build openapi

.PHONY: image
image: build
	docker build -t ghcr.io/Karaoke-Manager/server .

.PHONY: build
build: build/karman

.PHONY: openapi
openapi: build/openapi.html build/openapi.yaml

.PHONY: lint-openapi
lint-openapi: .redocly.yaml $(OPENAPI_FILES)
	@redocly lint $(OPENAPI_SPECS)

build/openapi.yaml: .redocly.yaml $(OPENAPI_FILES)
	@echo "Build OpenAPI Document"
	@redocly join --output "$@" $(OPENAPI_SPECS)
	@echo "Inserting Tag Groups"
	@yq -i '.x-tagGroups = ($(shell yq -j .x-tagGroups < api/karman.yaml))' "$@"

build/openapi.html: build/openapi.yaml
	@echo "Build HTML Documentation"
	@redocly build-docs --output "$@" "$<"

build/karman: $(GO_FILES)
	go build -o build/karman github.com/Karaoke-Manager/karman/cmd/karman
