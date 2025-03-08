#####################
#  Define Variables #
#####################

APP_ENV ?= dev

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

test:
	@go test ./internal/interface/transport/rest