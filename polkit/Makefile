# Polkit rule installation for Factorio Webapp
# This allows the factorio user to manage factorio@ systemd services

POLKIT_RULES_DIR=/etc/polkit-1/rules.d
RULE_FILE=50-factorio-webapp.rules

.PHONY: install uninstall help

# Install the Polkit rule
install:
	@echo "Installing Polkit rule to $(POLKIT_RULES_DIR)..."
	install -d $(POLKIT_RULES_DIR)
	install -m 644 $(RULE_FILE) $(POLKIT_RULES_DIR)/
	@echo "Reloading polkit..."
	@if command -v systemctl >/dev/null 2>&1; then \
		systemctl reload polkit 2>/dev/null || echo "Note: polkit service not found or reload not needed"; \
	else \
		echo "Note: systemctl not available, polkit may need manual reload"; \
	fi
	@echo "Polkit rule installed successfully!"
	@echo ""
	@echo "The factorio user can now manage factorio@ services."

# Uninstall the Polkit rule
uninstall:
	@echo "Removing Polkit rule from $(POLKIT_RULES_DIR)..."
	rm -f $(POLKIT_RULES_DIR)/$(RULE_FILE)
	@echo "Reloading polkit..."
	@if command -v systemctl >/dev/null 2>&1; then \
		systemctl reload polkit 2>/dev/null || echo "Note: polkit service not found or reload not needed"; \
	else \
		echo "Note: systemctl not available, polkit may need manual reload"; \
	fi
	@echo "Polkit rule removed successfully!"

# Show help
help:
	@echo "Factorio Webapp Polkit Rule Installation"
	@echo ""
	@echo "Usage:"
	@echo "  make install   - Install the Polkit rule (requires root)"
	@echo "  make uninstall - Remove the Polkit rule (requires root)"
	@echo "  make help      - Show this help message"
	@echo ""
	@echo "To install: sudo make install"
	@echo "To uninstall: sudo make uninstall"
