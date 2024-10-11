IMG ?= "ghcr.io/k8scope/k8s-restart-app:dev"
CLUSTER_NAME ?= "k8s-restart-app"

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} -f Dockerfile.dev .

##@ Setup

.PHONY: setup
kind-setup: ## Setup all the things for end-to-end tests in kind cluster
	$(MAKE) -C _e2e/ setup

.PHONY: kind-destroy
kind-destroy: ## Destroy the kind cluster
	$(MAKE) -C _e2e/ kind-destroy

.PHONY: helm-install
helm-install: ## Install all the necessary helm charts
	$(MAKE) -C _e2e/ helm-install

.PHONE: reload
reload: ## Add the docker image to the kind cluster and restart the restart-app
	$(MAKE) -C _e2e/ restart-restart-app

##@ End-to-End

.PHONY: e2e-run
e2e-run: ## Run end-to-end tests
	$(MAKE) -C _e2e/ e2e-run
