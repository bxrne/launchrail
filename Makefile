BINARY_NAME=launchrail
SRC_DIR=./cmd/launchrail
BUILD_DIR=./build

all: build

build:
	@echo "Building the project..."
	CGO_ENABLED=1 GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

build-windows:
	@echo "Installing Windows dependencies..."
	@go get -u github.com/go-gl/glfw/v3.3/glfw
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(SRC_DIR)

build-linux:
	@echo "Installing Linux dependencies..."
	@sudo apt-get install libgl1-mesa-dev libxcursor-dev libxrandr-dev libxinerama-dev -y
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

build-darwin:
	@echo "Installing macOS dependencies..."
	@brew install glew
	@go get -u github.com/go-gl/gl/v4.1-core/gl
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

test:
	@echo "Running tests..."
	go test ./... -coverprofile=coverage.out

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

dev:
	@echo "Running the project..."
	go run $(SRC_DIR)

.PHONY: all build build-windows build-linux build-darwin test clean dev


