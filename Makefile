IMG ?= "ghcr.io/k8scope/k8s-restart-app:dev"

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} -f Dockerfile.dev .
