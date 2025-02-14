# Variables
CLUSTER_NAME = dice-cluster
APP_NAME = dice-app
IMAGE_TAG = v2
GRAFANA_ENDPOINT = htt/otlp
GRAFANA_USERNAME = ""
GRAFANA_PASSWORD = ""

.PHONY: all
all: cluster build load deploy

.PHONY: cluster
cluster:
	@echo "Creating Kind cluster..."
	kind create cluster --name $(CLUSTER_NAME) --config kind.yaml || true

.PHONY: pre-requisites
pre-requisites:
	@echo "Installing Ingress Nginx Controller..."
	kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml
	@echo "Creating o11y namespace..."
	kubectl create namespace o11y
	@echo "Creating credentials secret..."
	kubectl create secret generic gcloud-creds --from-literal=endpoint=$(GRAFANA_ENDPOINT) --from-literal=username=$(GRAFANA_USERNAME) --from-literal=password=$(GRAFANA_PASSWORD) --namespace=o11y 

.PHONY: build
build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(IMAGE_TAG) .

.PHONY: load
load:
	@echo "Loading image into Kind cluster..."
	kind load docker-image $(APP_NAME):$(IMAGE_TAG) --name $(CLUSTER_NAME)

.PHONY: deploy-alloy
deploy-alloy:
	@echo "Deploying to cluster..."
	helm install alloy-service -f deploy/values-alloy.yaml grafana/alloy --namespace o11y --create-namespace

.PHONY: rm-alloy	
rm-alloy:
	@echo "Removing alloy service..."
	helm delete alloy-service --namespace o11y

.PHONY: deploy-dice
deploy-dice:
	@echo "Deploying to cluster..."
	kubectl apply -f deploy/deployment.yaml --namespace o11y

.PHONY: clean
clean:
	@echo "Cleaning up..."
	kind delete cluster --name $(CLUSTER_NAME)

.PHONY: delete
delete:
	@echo "Deleting deployment..."
	helm delete alloy-service --namespace o11y
