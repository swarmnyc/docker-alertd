.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build docker image
	@echo "building docker image"
	@docker build -t swarmnyc/docker-alertd .

push: ## Push image to swarmnyc/docker-alertd
	@echo "pushing to docker registry"
	@docker push swarmnyc/docker-alertd