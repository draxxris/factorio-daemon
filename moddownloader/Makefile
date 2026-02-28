BINARY_NAME=moddownloader
BUILD_DIR=../build
MAIN_PACKAGE=./cmd/main

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

test:
	$(GOTEST) -v -cover ./...

test-integration:
	bash -c "source ../.env && $(GOTEST) -v -tags=integration ./..."

deps:
	$(GOGET) -d -v ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: build clean test deps run
