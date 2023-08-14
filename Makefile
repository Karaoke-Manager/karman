OPENAPI_FILES := $(shell find ./openapi/ -type f -name '*.yaml')
OPENAPI_SPECS := ./openapi/karman.yaml $(shell find ./openapi/tags/ -type f -name '*.yaml')
GO_FILES = $(shell find . -type f -name '*.go')

.PHONY: openapi
openapi: build/openapi.html build/openapi.yaml

.PHONY: lint-openapi
lint-openapi: .redocly.yaml $(OPENAPI_FILES)
	@redocly lint $(OPENAPI_SPECS)

build/openapi.yaml: .redocly.yaml $(OPENAPI_FILES)
	@echo "Build OpenAPI Document"
	@redocly join --output "$@" $(OPENAPI_SPECS)
	@echo "Inserting Tag Groups"
	@yq -i '.x-tagGroups = ($(shell yq -j .x-tagGroups < ./openapi/karman.yaml))' "$@"

build/openapi.html: build/openapi.yaml
	@echo "Build HTML Documentation"
	@redocly build-docs --output "$@" "$<"
