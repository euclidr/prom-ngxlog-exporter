RELEASE_DIR = dist
TARGETS = linux-amd64 darwin-amd64
VERSION = v0.1.0
APP_NAME = ngxlog-exporter

OBJECTS = $(patsubst %,$(RELEASE_DIR)/$(APP_NAME)-%-$(VERSION), $(TARGETS))

LDFLAGS = -ldflags "-s -w"

.PHONY: build
build: $(OBJECTS) ## Build excutables
	@echo "\x1b[32mbuilding excutables......\x1b[0m"

$(OBJECTS): $(wildcard *.go)
	env GOOS=`echo $@ | cut -d'-' -f3` GOARCH=`echo $@ | cut -d'-' -f4` go build -o $@ $(LDFLAGS) .

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
