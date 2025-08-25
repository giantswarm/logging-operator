.PHONY: update-golden-files
update-golden-files: ## Update golden files.
	@echo "Updating golden files..."
	@go test -v \
		./pkg/resource/logging-config \
		./pkg/resource/events-logger-config \
		-update
