# Include variables from the .envrc file
-include .envrc

## help - print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== # 
# DEVELOPMENT
# ==================================================================================== #

## run - runs the dkg node
.PHONY: run
run: build
	@echo "Starting dkg service..."
	adkg start
	
# ==================================================================================== # 
# BUILD
# ==================================================================================== #

## build - installs all depdencies and builds go binary in GOPATH
.PHONY: build
build:
	@echo 'Building binary...'
	@go install main.go -o adkg
	
# ==================================================================================== # 
# QUALITY CONTROL
# ==================================================================================== #

## lint - runs linter across the project using golangci-lint
.PHONY: lint
lint: 
	@golangci-lint run  ./...

## upgrade - upgrades all go dependencies
.PHONY: upgrade
upgrade:
	@echo "Upgrading dependencies..."
	@go get -u
	@go mod tidy

## test - runs the test file across the project
.PHONY: test
test:
	@echo 'Running tests and reporting to coverage.txt...'
	@go test -coverprofile=coverage.txt ./...

## show-container-logs - shows the logs for DKG nodes
.PHONY: show-container-logs
show-container-logs:
	@docker-compose logs -f

## run-service - runs DKG nodes
.PHONY: run-service
run-service:
	@docker-compose up -d

## clean - cleans everything
.PHONY: clean
clean:
	@echo "Cleaning..."
	@docker-compose down
	@echo "Cleaned"

## restart - restarts DKG nodes
.PHONY: restart
restart:
	@echo "Restarting..."
	@docker-compose restart
	@echo "Done"

.PHONY: run-local
run-local:
	@make run-service

.PHONY: build-docker-image
build-docker-image:
	@docker build . -t dkgnode