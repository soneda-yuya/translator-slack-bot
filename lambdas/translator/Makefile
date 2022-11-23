.DEFAULT_GOAL := help

.PHONY: build
build: ## build
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./build/main

.PHONY: help
help:
	@grep -E '^[\/a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'