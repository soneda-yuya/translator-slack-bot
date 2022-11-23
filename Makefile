.DEFAULT_GOAL := help

PWD = $(shell pwd)

TERRAFORM_VERSION=1.3.3
RUN_TERRAFORM=docker run \
	--rm \
	--volume "$(PWD):/workspace" \
	--env AWS_ACCESS_KEY_ID='$(shell echo $$AWS_ACCESS_KEY_ID)' \
	--env AWS_SECRET_ACCESS_KEY='$(shell echo $$AWS_SECRET_ACCESS_KEY)' \
	--env AWS_SESSION_TOKEN='$(shell echo $$AWS_SESSION_TOKEN)' \
	--workdir /workspace hashicorp/terraform:$(TERRAFORM_VERSION)

SERVICES := translator
ENVIRONMENTS := stage
SERVICES := translator
COMMANDS := init plan apply output validate

ACTION_TARGETS := $(foreach C, $(ENVIRONMENTS), $(addprefix $(C)/, $(SERVICES)))
TARGETS := $(foreach C, $(ACTION_TARGETS), $(addprefix $(C)/, $(COMMANDS)))

.PHONY: deps
deps: _vendor/bin/direnv env ## ready dependency

.PHONY: commands
commands: ## commands
	@$(foreach name, $(TARGETS), echo $(name);)

.PHONY: env
env: ## load direnv
	direnv allow

.PHONY: $(TARGETS)
$(TARGETS):
	@$(RUN_TERRAFORM) -chdir=terraform/enviroments/$(@D) $(@F) $(OPT)

.PHONY: fmt
fmt: ## fmt
	@$(RUN_TERRAFORM) fmt -recursive

.PHONY: help
help:
	@grep -E '^[\/a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'