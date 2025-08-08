-include .env
export

.PHONY: buf.gen run-product docker-compose-start docker-compose-stop wire.gen

# os : linux
PROJECT_DIR := $(shell pwd)

buf.gen:
	docker run --rm --volume "$(PROJECT_DIR):/workspace" --workdir /workspace bufbuild/buf generate

wire.gen:
	docker run --rm --volume "$(PROJECT_DIR):/workspace" --workdir /workspace/internal/product/app golang:1.21-alpine sh -c "go install github.com/google/wire/cmd/wire@latest && /go/bin/wire"

run-product:
	cd cmd/product && go mod tidy && go mod download && \
	CGO_ENABLED=0 go run main.go

docker-compose-start:
	docker compose up --build

docker-compose-stop:
	docker compose down --remove-orphans -v