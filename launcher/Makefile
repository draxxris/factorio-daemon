# Installation settings
PREFIX=/opt/factorio
SERVICE_FILE=/etc/systemd/system/factorio@.service

# Remote deployment settings
#REMOTE_HOST=deploy@192.168.1.30
#REMOTE_PATH=/opt/factorio/

# Files to install
SCRIPTS=factorio.sh
TEMPLATES=templates/server-settings.json templates/config.ini
EXAMPLE_ENV=env-example

# Mod downloader tool
MODDOWNLOADER_DIR=../moddownloader

.PHONY: install clean deploy moddownloader test

# Build moddownloader
moddownloader:
	$(MAKE) -C $(MODDOWNLOADER_DIR) build

# Install the launcher
install: moddownloader
	@echo "Installing factorio launcher to $(PREFIX)..."
	@echo "Creating factorio system user..."
	@id factorio >/dev/null 2>&1 || \
		useradd --system --home-dir $(PREFIX) --shell /usr/sbin/nologin factorio
	install -d $(PREFIX)
	install -d $(PREFIX)/templates
	install -m 755 $(SCRIPTS) $(PREFIX)/
	install -m 755 ../build/moddownloader $(PREFIX)/
	install -m 644 $(TEMPLATES) $(PREFIX)/templates/
	install -m 644 $(EXAMPLE_ENV) $(PREFIX)/
	install -m 644 factorio@.service $(SERVICE_FILE)
	@echo "Setting ownership and permissions..."
	chown -R factorio:factorio $(PREFIX)
	chmod 750 $(PREFIX)
	@echo "Installation complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Create a .env file with your Factorio credentials:"
	@echo "   echo 'FACTORIO_USERNAME=your_username' > $(PREFIX)/.env"
	@echo "   echo 'FACTORIO_TOKEN=your_token' >> $(PREFIX)/.env"
	@echo "   sudo chown factorio:factorio $(PREFIX)/.env"
	@echo "   sudo chmod 640 $(PREFIX)/.env"
	@echo ""
	@echo "2. Create an instance configuration:"
	@echo "   cp $(PREFIX)/env-example $(PREFIX)/env-myserver"
	@echo "   sudo nano $(PREFIX)/env-myserver"
	@echo "   sudo chown factorio:factorio $(PREFIX)/env-myserver"
	@echo "   sudo chmod 640 $(PREFIX)/env-myserver"
	@echo ""
	@echo "3. Install Polkit rule for webapp service management:"
	@echo "   sudo install -d /etc/polkit-1/rules.d"
	@echo "   sudo install -m 644 ../polkit/50-factorio-webapp.rules /etc/polkit-1/rules.d/"
	@echo ""
	@echo "4. Reload systemd and start your instance:"
	@echo "   sudo systemctl daemon-reload"
	@echo "   sudo systemctl start factorio@myserver"
	@echo "   sudo systemctl enable factorio@myserver"

# Deploy to remote server via SCP
deploy: moddownloader
	ssh $(REMOTE_HOST) mkdir -p $(REMOTE_PATH)/launcher
	scp $(SCRIPTS) factorio@.service $(REMOTE_HOST):$(REMOTE_PATH)/launcher
	scp ../build/moddownloader $(REMOTE_HOST):$(REMOTE_PATH)/launcher
	scp -r templates $(REMOTE_HOST):$(REMOTE_PATH)/launcher
	scp env-* $(REMOTE_HOST):$(REMOTE_PATH)/
	@echo "Deployment complete!"

clean:
	@echo "Nothing to clean in launcher directory"
	$(MAKE) -C $(MODDOWNLOADER_DIR) clean

# Run tests (shellcheck for scripts)
test:
	@echo "Running shellcheck on scripts..."
	@command -v shellcheck >/dev/null 2>&1 || { echo "Error: shellcheck not installed."; exit 1; }
	@shellcheck factorio.sh && echo "✓ shellcheck passed" || { echo "✗ shellcheck failed"; exit 1; }
	@echo "All tests passed!"
