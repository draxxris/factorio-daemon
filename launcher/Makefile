# Installation settings
PREFIX=/opt/factorio
SERVICE_FILE=/etc/systemd/system/factorio@.service

# Remote deployment settings
#REMOTE_HOST=deploy@192.168.1.30
#REMOTE_PATH=/opt/factorio/

# Files to install
SCRIPTS=factorio.sh download_mods.py
TEMPLATES=templates/server-settings.json templates/config.ini
EXAMPLE_ENV=env-example

.PHONY: install clean deploy

# Install the launcher
install:
	@echo "Installing factorio launcher to $(PREFIX)..."
	install -d $(PREFIX)
	install -d $(PREFIX)/templates
	install -m 755 $(SCRIPTS) $(PREFIX)/
	install -m 644 $(TEMPLATES) $(PREFIX)/templates/
	install -m 644 $(EXAMPLE_ENV) $(PREFIX)/
	install -m 644 factorio@.service $(SERVICE_FILE)
	@echo "Installation complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Create a .env file with your Factorio credentials:"
	@echo "   echo 'FACTORIO_USERNAME=your_username' > $(PREFIX)/.env"
	@echo "   echo 'FACTORIO_TOKEN=your_token' >> $(PREFIX)/.env"
	@echo ""
	@echo "2. Create an instance configuration:"
	@echo "   cp $(PREFIX)/env-example $(PREFIX)/env-myserver"
	@echo "   nano $(PREFIX)/env-myserver"
	@echo ""
	@echo "3. Reload systemd and start your instance:"
	@echo "   sudo systemctl daemon-reload"
	@echo "   sudo systemctl start factorio@myserver"
	@echo "   sudo systemctl enable factorio@myserver"

# Deploy to remote server via SCP
deploy:
	ssh $(REMOTE_HOST) mkdir -p $(REMOTE_PATH)/launcher
	scp $(SCRIPTS) factorio@.service $(REMOTE_HOST):$(REMOTE_PATH)/launcher
	scp -r templates $(REMOTE_HOST):$(REMOTE_PATH)/launcher
	scp env-* $(EXAMPLE_ENV) $(REMOTE_HOST):$(REMOTE_PATH)/
	@echo "Deployment complete!"

clean:
	@echo "Nothing to clean in launcher directory"
