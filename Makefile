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

# Unit Test
test:
	@go test ./internal/interface/transport/rest

# Integration Test
it:
	@go build -cover -a -o e-lib-it github.com/DarrelA/e-lib/cmd/loan
	./deployment/scripts/wrap_test_for_coverage.sh
	@go tool cover -html=./testdata/reports/covdatafiles/coverage.out -o "./testdata/reports/it_coverage.html"