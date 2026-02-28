BINARY_NAME=factorio-webapp
BUILD_DIR=../build
MAIN_PACKAGE=./cmd/server

# Installation settings
PREFIX=/opt/factorio
INSTALL_DIR=$(PREFIX)/webapp
SERVICE_FILE=/etc/systemd/system/factorio-webapp.service

# Remote deployment settings
#REMOTE_HOST=deploy@192.168.1.30
#REMOTE_PATH=/opt/factorio/

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Build the webapp
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

test:
	$(GOTEST) -v -cover ./...

# Install the webapp
install: build
	@echo "Installing factorio-webapp to $(INSTALL_DIR)..."
	install -d $(INSTALL_DIR)
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	install -d $(INSTALL_DIR)/data/staging
	install -d $(INSTALL_DIR)/data/backups
	@echo "Installing systemd service file..."
	install -m 644 factorio-webapp.service $(SERVICE_FILE)
	@echo "Installation complete!"
	@echo "Run: sudo systemctl daemon-reload"
	@echo "Run: sudo systemctl enable --now factorio-webapp"

deps:
	$(GOGET) -d -v ./...

# Developer tasks
## Deploy to remote server via SCP
deploy: build
	scp $(BUILD_DIR)/$(BINARY_NAME) $(REMOTE_HOST):$(REMOTE_PATH)/webapp/
## Run locally for testing
run: build
	./$(BINARY_NAME)

.PHONY: build clean test deploy deps run install
