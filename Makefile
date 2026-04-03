BINARY = cc-dash
VERSION = 0.1.0
BUILD_DIR = build
FRONTEND_DIR = frontend
STATIC_DIR = internal/server/static

.PHONY: dev build build-frontend build-backend build-all clean install run stop

# Desenvolvimento: roda Go server (precisa build-frontend antes)
dev: build-frontend
	go run ./cmd/cc-dash --port 3000

# Desenvolvimento frontend (com proxy para Go)
dev-frontend:
	cd $(FRONTEND_DIR) && bun run dev

# Build completo: frontend + backend
build: build-frontend build-backend

# Build do frontend e copia para static/
build-frontend:
	@echo ">> building frontend..."
	@cd $(FRONTEND_DIR) && bun install --frozen-lockfile 2>/dev/null || bun install && bun run build
	@mkdir -p $(STATIC_DIR)/assets
	@cp $(FRONTEND_DIR)/dist/index.html $(STATIC_DIR)/
	@cp -r $(FRONTEND_DIR)/dist/assets/* $(STATIC_DIR)/assets/
	@echo ">> frontend pronto"

# Build do backend
build-backend:
	@echo ">> building $(BINARY) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) ./cmd/cc-dash
	@echo ">> $(BUILD_DIR)/$(BINARY) ($$(du -h $(BUILD_DIR)/$(BINARY) | cut -f1))"

# Build e rodar
run: build
	./$(BUILD_DIR)/$(BINARY)

# Cross-compilation
build-all: build-frontend
	@echo ">> cross-compiling..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/cc-dash
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/cc-dash
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/cc-dash
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/cc-dash
	@echo ">> builds:"
	@ls -lh $(BUILD_DIR)/$(BINARY)-*

# Instalar localmente
install: build
	@cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY) 2>/dev/null || \
		cp $(BUILD_DIR)/$(BINARY) $$HOME/go/bin/$(BINARY) 2>/dev/null || \
		echo "Copie manualmente: cp $(BUILD_DIR)/$(BINARY) para um diretorio no PATH"
	@echo ">> $(BINARY) instalado"

# Stop all cc-dash processes
stop:
	@pids=$$(pgrep -f '$(BINARY)' | grep -v $$$$); \
	if [ -n "$$pids" ]; then \
		echo ">> stopping cc-dash (PIDs: $$pids)"; \
		echo "$$pids" | xargs kill 2>/dev/null; \
		sleep 1; \
		alive=$$(pgrep -f '$(BINARY)' | grep -v $$$$); \
		if [ -n "$$alive" ]; then \
			echo ">> force killing (PIDs: $$alive)"; \
			echo "$$alive" | xargs kill -9 2>/dev/null; \
		fi; \
		echo ">> stopped"; \
	else \
		echo ">> no cc-dash processes running"; \
	fi

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) $(BINARY)
	@echo ">> cleaned"
