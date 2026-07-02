PROJECT_NAME := "alfred-github-search"
PKG          := "github.com/rwilgaard/$(PROJECT_NAME)"
GO111MODULE  = on

.EXPORT_ALL_VARIABLES:
.PHONY: all dep fmt lint vet build clean universal-binary package-alfred release help

all: build

dep: ## Get the dependencies
	@go mod download

fmt: ## Format Go files with gofumpt
	@gofumpt -l -w ./src

lint: ## Lint Golang files
	@golangci-lint run --timeout 3m

vet: ## Run go vet
	@go vet ./src

build: dep ## Build arch-specific binaries
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o workflow/$(PROJECT_NAME)-amd64 ./src
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o workflow/$(PROJECT_NAME)-arm64 ./src

universal-binary: ## Combine arch binaries into universal binary
	@lipo -create -output workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64
	@rm -f workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

clean: ## Remove build artifacts
	@rm -f workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

package-alfred: build universal-binary ## Build and package into .alfredworkflow
	@cd ./workflow && zip -r ../$(PROJECT_NAME).alfredworkflow ./*
	@rm -f workflow/$(PROJECT_NAME)
	@echo "Created $(PROJECT_NAME).alfredworkflow"

release: ## Prepare and tag a new release (usage: make release VERSION=x.y.z)
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=x.y.z"; exit 1; fi
	@plutil -replace version -string "$(VERSION)" workflow/info.plist
	@make package-alfred
	@git add workflow/info.plist $(PROJECT_NAME).alfredworkflow
	@git commit -m "chore: release v$(VERSION)"
	@git tag "v$(VERSION)"
	@git push origin main --tags
	@echo "Released v$(VERSION)"

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
