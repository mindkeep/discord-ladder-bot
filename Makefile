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
CONFIG_FILE := config.yml
SOURCES := $(shell find . -name '*.go')
VERSION_FILE := VERSION

# Default action
default: test build push

bin:
	mkdir -p bin

# Build, vet, and test Go code
bin/$(BIN_NAME): bin $(SOURCES)
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/$(BIN_NAME) cmd/main/main.go

# vet and test
test:
	go vet ./...
	go test ./...

# Increment the build version
increment-version:
	@echo "Incrementing build version..."
	@VERSION=$$(cat $(VERSION_FILE)); \
 	NEW_VERSION=$$(echo $$VERSION | awk -F. '{print $$1"."$$2"."$$3+1}'); \
	echo $$NEW_VERSION > $(VERSION_FILE); \
	echo "New version: $$NEW_VERSION"

# Build the image using Buildah
build: bin/$(BIN_NAME) increment-version
	@VERSION=$$(cat $(VERSION_FILE)); \
	buildah bud -f $(DOCKERFILE) -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) -t $(REGISTRY)/$(IMAGE_NAME):$$VERSION .

# Push the image to a registry
push: build
	@VERSION=$$(cat $(VERSION_FILE)); \
	podman push $(REGISTRY)/$(IMAGE_NAME):$(TAG); \
	podman push $(REGISTRY)/$(IMAGE_NAME):$$VERSION

# Run the container using Podman
podman-run: push
	podman run --rm -it --secret $(SECRET_ID),target=/app/config.yml $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Clean up local images and Go binaries
clean:
	podman rmi $(REGISTRY)/$(IMAGE_NAME):$(TAG)
	@VERSION=$$(cat $(VERSION_FILE)); \
	podman rmi $(REGISTRY)/$(IMAGE_NAME):$$VERSION
	rm -f bin/$(BIN_NAME)

.PHONY: default build run push deploy clean go-build go-vet go-test increment-version
