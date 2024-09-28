# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
SWAG = swag

# Name of binary
BINARY_NAME = phoenix-go
BUILD_DIR = bin

# Platforms for cross-compilation
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# Build the binary
build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go

# Clean build files
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOCLEAN) -testcache
	$(GOTEST) -v ./...

# Cross-compile for all specified platforms and zip each binary
build-all: clean
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d '/' -f 1); \
		GOARCH=$$(echo $$platform | cut -d '/' -f 2); \
		OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) -ldflags "-s -w" -o $$OUTPUT cmd/main.go || exit 1; \
		echo "Built: $$OUTPUT"; \
		zip -j $$OUTPUT.zip $$OUTPUT || exit 1; \
		echo "Zipped: $$OUTPUT.zip"; \
	done

# Generate Swagger documentation
swagger:
	$(SWAG) init -g cmd/main.go --output ./docs

# Default target
default: build

.PHONY: build clean test build-all swagger
