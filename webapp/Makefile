BINARY_NAME=factorio-webapp
BUILD_DIR=../build
MAIN_PACKAGE=./cmd/server

# Installation settings
PREFIX=/opt/factorio
INSTALL_DIR=$(PREFIX)/webapp
SERVICE_FILE=/etc/systemd/system/factorio-webapp.service

# Sudoers settings
SUDOERS_DIR=/etc/sudoers.d
SUDOERS_RULE_FILE=sudoers/factorio-webapp

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
	@echo "Setting ownership and permissions..."
	chown -R factorio:factorio $(INSTALL_DIR)
	chmod 750 $(INSTALL_DIR)
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

## Setup nginx with basic authentication
nginx-setup:
	@echo "Setting up nginx configuration for Factorio Webapp..."
	@echo ""
	@echo "This will configure nginx as a reverse proxy with basic authentication."
	@echo "You will be prompted for a username and password."
	@echo ""
	@read -p "Enter username for webapp access: " username; \
	if [ -z "$$username" ]; then \
		echo "Error: Username cannot be empty"; \
		exit 1; \
	fi; \
	stty -echo; \
	read -p "Enter password for webapp access: " password; \
	stty echo; \
	echo ""; \
	if [ -z "$$password" ]; then \
		echo "Error: Password cannot be empty"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "Installing nginx configuration..."; \
	install -d /etc/nginx/conf.d; \
	install -m 644 nginx.conf.template /etc/nginx/conf.d/factorio.conf; \
	echo "Creating htpasswd file..."; \
	which htpasswd >/dev/null 2>&1 || { \
		echo "Error: htpasswd not found. Please install apache2-utils (Debian/Ubuntu) or httpd-tools (RHEL/CentOS)"; \
		exit 1; \
	}; \
	htpasswd -cb /etc/nginx/.htpasswd-factorio "$$username" "$$password"; \
	chmod 640 /etc/nginx/.htpasswd-factorio; \
	echo ""; \
	echo "Testing nginx configuration..."; \
	nginx -t 2>/dev/null || { \
		echo "Error: nginx configuration test failed"; \
		exit 1; \
	}; \
	echo "Reloading nginx..."; \
	systemctl reload nginx || systemctl restart nginx; \
	echo ""; \
	echo "============================================"; \
	echo "nginx setup complete!"; \
	echo "============================================"; \
	echo ""; \
	echo "Access the webapp at: http://your-server/"; \
	echo "Username: $$username"; \
	echo ""; \
	echo "To remove this configuration later, run: sudo rm /etc/nginx/conf.d/factorio.conf /etc/nginx/.htpasswd-factorio && sudo systemctl reload nginx"

.PHONY: build clean test deploy deps run install nginx-setup install-sudoers uninstall-sudoers

# Install sudoers rule for factorio user
install-sudoers:
	@echo "Installing sudoers rule to $(SUDOERS_DIR)..."
	@echo "WARNING: This allows the factorio user to manage factorio@ services without password."
	@echo "         Only install if you trust the factorio user and webapp."
	@echo ""
	install -d $(SUDOERS_DIR)
	install -m 440 $(SUDOERS_RULE_FILE) $(SUDOERS_DIR)/
	@echo "Verifying sudoers configuration..."
	@visudo -c $(SUDOERS_DIR)/factorio-webapp || { \
		echo "Error: sudoers configuration is invalid!"; \
		rm -f $(SUDOERS_DIR)/factorio-webapp; \
		exit 1; \
	}
	@echo "Sudoers rule installed successfully!"
	@echo ""
	@echo "The factorio user can now manage factorio@ services without password."

# Uninstall sudoers rule
uninstall-sudoers:
	@echo "Removing sudoers rule from $(SUDOERS_DIR)..."
	rm -f $(SUDOERS_DIR)/factorio-webapp
	@echo "Sudoers rule removed successfully!"
