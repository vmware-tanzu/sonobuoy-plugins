REGISTRY = "projects.registry.vmware.com/cre"
IMAGE = "reliability-scanner"
TAG = "v0.1.1"

.SILENT: cycle
clean:
	rm -rf ./generated/
	rm -rf *.tar.gz
	-kubectl delete ns sonobuoy
docker-build:
	docker build -t ${REGISTRY}/${IMAGE}:${TAG} .
docker-push:
	docker push ${REGISTRY}/${IMAGE}:${TAG}
run:
	ytt -f plugin/default-ytt.yaml -f  plugin/reliability-scanner-custom-values.lib.yml | kubectl apply -f-
	@./hack/waiter.sh
results:
	$(eval OUT=$(shell sonobuoy retrieve))
	sonobuoy results --mode detailed --plugin reliability-scanner $(OUT)
generate: # if changing PLUGIN_BASE content. New Sonobuoy manifests will need to be generated. Please preserve ./plugin/default.yaml ytt metadata. 
	rm -rf ./generated/ && mkdir -p generated && \
	sonobuoy gen --plugin ./plugin/base.yaml --show-default-podspec > ./generated/default.yaml
cycle: clean docker-build docker-push run
