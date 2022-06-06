SA := $(shell for f in ./manifest/*; do cat $$f | yq '(.|select(.kind == "ServiceAccount")).metadata.name' ; done)
IMAGE := $(shell for f in ./manifest/*; do cat $$f | yq "(.|select(.spec.template.spec.serviceAccountName == \"$(SA)\")).spec.template.spec.containers[].image" ; done)
kind_name ?= kind


prepare:
	echo $(SA)
	echo $(IMAGE)
	docker build -t $(IMAGE) .
	kind load docker-image $(IMAGE) --name $(kind_name)

run: prepare
	kubectl apply -f manifest
	@bash .validate/helper.sh

all: run
	 @bash .validate/validate-rule1.sh
	 @bash .validate/validate-rule2.sh
	 @bash .validate/validate-rule3.sh
	 @bash .validate/validate-rule4.sh



