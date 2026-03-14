SPEC_URL := https://raw.githubusercontent.com/frontapp/front-api-specs/main/core-api/core-api.json

.PHONY: build test vet generate

build:
	go build -o front ./cmd/front

test:
	go test ./...

vet:
	go vet ./...

generate:
	curl -fsSL $(SPEC_URL) -o internal/api/core-api.json
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config internal/api/cfg.yaml internal/api/core-api.json
	gofmt -w internal/api/front.gen.go
