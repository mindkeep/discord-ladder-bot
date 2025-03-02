# Load environment variables from .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Variables
IMAGE_NAME := discord-ladder-bot
BIN_NAME := discord-ladder-bot
TAG := latest
DOCKERFILE := Dockerfile
CONFIG_FILE := config.yaml
SOURCES := $(shell find . -name '*.go')

# Default action
default: test build

bin:
	mkdir bin

# Build, vet, and test Go code
bin/$(BIN_NAME): bin $(SOURCES)
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/$(BIN_NAME) cmd/main/main.go

# vet and test
test:
	go vet ./...
	go test ./...

# Build the image using Buildah
build: bin/$(BIN_NAME)
	buildah bud -f $(DOCKERFILE) -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) .


# Push the image to a registry
push: build
	podman push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Run the container using Podman
podman-run: push
	podman run --rm -it --secret $(SECRET_ID),target=/app/config.yml $(REGISTRY)/$(IMAGE_NAME):$(TAG)

deploy: push
	kubectl apply -f deployment.yml

# Clean up local images and Go binaries
clean:
	podman rmi $(REGISTRY)/$(IMAGE_NAME):$(TAG)
	rm -rf bin

.PHONY: default test build podman-run push deploy clean
