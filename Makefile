#####################
#  Define Variables #
#####################

APP_ENV ?= dev
VARS = POSTGRES_DB=$(POSTGRES_DB) POSTGRES_USER=$(POSTGRES_USER) 

#####################
#    Env Configs    #
#####################

# Path to the environment-specific .env file
ENV_FILE=./config/.env.$(APP_ENV)

# Check if the environment-specific .env file exists
ifeq (,$(wildcard $(ENV_FILE)))
  $(error "$(ENV_FILE) file not found")
endif

# Include environment-specific variables from .env.${APP_ENV}
include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE))

#####################
#    make <cmd>     #
#####################

dev:
	@go run cmd/loan/main.go

# Unit Test
test:
	@go test ./internal/interface/transport/rest

# Integration Test
# Depends on: deployment/docker-compose.integration.yml
.PHONY: it

it:
	@echo "Running integration tests with docker-compose.integration.yml..."
	@cd deployment && \
		APP_ENV=test $(VARS) docker compose -f docker-compose.integration.yml build app-integration-test && \
		APP_ENV=test $(VARS) docker compose -f docker-compose.integration.yml run --rm app-integration-test
	@go tool cover -html=./testdata/reports/covdatafiles/coverage.out -o "./testdata/reports/it_coverage.html"
	@echo "Removing containers and volumes..."
	@cd deployment && APP_ENV=test $(VARS) docker compose -f docker-compose.integration.yml down -v
