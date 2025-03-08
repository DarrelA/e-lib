#####################
#    make <cmd>     #
#####################

dev:
	@go run cmd/loan/main.go

test:
	@go test ./internal/interface/transport/rest