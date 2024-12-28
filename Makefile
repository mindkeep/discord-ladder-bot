# Load environment variables from .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Variables
IMAGE_NAME := discord-ladder-bot
TAG := latest
DOCKERFILE := Dockerfile
CONFIG_FILE := config.yaml

# Default action
default: build push run

# Build, vet, and test Go code
go-build:
	CGO_ENABLED=0 GOOS=linux go build -a -o app cmd/main/main.go

go-vet:
	go vet ./...

go-test:
	go test ./...

# Build the image using Buildah
build: go-build go-vet go-test
	buildah bud -f $(DOCKERFILE) -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) .

# Run the container using Podman
run: build
	podman run --rm -it --secret $(SECRET_ID),target=/app/config.yml $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Push the image to a registry
push: build
	podman push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

deploy: push
	kubectl apply -f deployment.yaml

# Clean up local images and Go binaries
clean:
	podman rmi $(REGISTRY)/$(IMAGE_NAME):$(TAG)
	rm -f app

.PHONY: default build run push deploy clean go-build go-vet go-test