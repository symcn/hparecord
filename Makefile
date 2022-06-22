IMAGE_REGISTRY_NAME?=symcn.tencentcloudcr.com/symcn
IMAGE_NAME?=hparecord
IMAGE_TAG?=v0.1.2
NAMESPACE?=sym-admin

docker-build:
	docker build \
	-f Dockerfile \
	-t $(IMAGE_REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE_REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)

deploy:
	helm upgrade --install --force hparecord --namespace ${NAMESPACE} --set image.tag=${IMAGE_TAG},image.pullPolicy="Always" ./charts/hparecord
