# os : linux
PROJECT_DIR := $(shell pwd)

buf.gen:
	docker run --rm --volume "$(PROJECT_DIR):/workspace" --workdir /workspace bufbuild/buf generate

docker-compose-start:
	docker compose up --build

docker-compose-stop:
	docker compose down --remove-orphans -v