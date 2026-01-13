.PHONY: build clean test run

GO := /usr/local/go/bin/go
BINARY := sentinel
CGO_LDFLAGS := -L/usr/lib

build:
	@echo "Building $(BINARY)..."
	CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) build -o $(BINARY) .
	@echo "Build complete: $(BINARY)"

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -f $(BINARY)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	$(GO) test ./...

run: build
	@echo "Running $(BINARY)..."
	./$(BINARY)

daemon: build
	@echo "Starting $(BINARY) as daemon..."
	./$(BINARY) -daemon -action start

stop:
	@echo "Stopping $(BINARY)..."
	./$(BINARY) -action stop

status:
	@echo "Checking $(BINARY) status..."
	./$(BINARY) -action status

mod-tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

help:
	@echo "Available targets:"
	@echo "  build       - Build the project"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  run         - Build and run in foreground"
	@echo "  daemon      - Build and start as daemon"
	@echo "  stop        - Stop the daemon"
	@echo "  status      - Check daemon status"
	@echo "  mod-tidy    - Tidy Go dependencies"
