IMG ?= "ghcr.io/k8scope/k8s-restart-app:dev"
CLUSTER_NAME ?= "k8s-restart-app"
RESTART_APP_NAMESPACE ?= "restart-app-system"

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

.PHONY: kind-setup
kind-setup: ## Create a kind cluster
	kind create cluster --name ${CLUSTER_NAME} --config config/kind_config.yaml

.PHONY: kind-destroy
kind-destroy: ## Destroy the kind cluster
	kind delete cluster --name ${CLUSTER_NAME}

.PHONY: kind-load
kind-load: ## Load docker image into kind cluster
	kind load docker-image ${IMG} --name ${CLUSTER_NAME}

.PHONY: helm-repo-setup
helm-repo-setup: ## Add the necessary helm repos
	helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx --force-update
	helm repo add k8scope https://k8scope.github.io/helm-charts --force-update
	helm repo update

.PHONY: helm-install-nginx
helm-install-nginx: ## Install the nginx ingress controller
	helm upgrade --install nginx ingress-nginx/ingress-nginx \
		-n ingress-system \
		--create-namespace \
		--set controller.service.type=NodePort \
		--set controller.service.nodePorts.http=31080 \
		--set controller.service.nodePorts.https=31443 \
		--set controller.admissionWebhooks.enabled=false

.PHONY: helm-install-restart-app
helm-install-restart-app: ## Install the restart-app helm chart
	helm upgrade --install restart-app k8scope/restart-controller \
		-n ${RESTART_APP_NAMESPACE} \
		--create-namespace \
		-f config/restart_app_config.yaml

.PHONY: install-test-apps
install-test-apps: ## Install the test apps
	kubectl apply -f config/apps.yaml

.PHONY: helm-install
helm-install: kind-load install-test-apps helm-repo-setup helm-install-nginx helm-install-restart-app ## Install all the necessary helm charts

.PHONY: setup
setup: kind-setup kind-load helm-repo-setup helm-install ## Setup the environment from scratch

.PHONY: restart-restart-app
restart-restart-app: kind-load ## Restart the restart-app
	kubectl rollout restart deployment restart-app-restart-controller -n ${RESTART_APP_NAMESPACE}

##@ End-to-End

.PHONY: e2e-run
e2e-run: ## Run end-to-end tests
	SERVICE_ADDRESS="http://restart-app.localhost" go test -v ./ -coverprofile=coverage.txt -covermode=atomic
