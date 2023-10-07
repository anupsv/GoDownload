# Name of the output binary
BINARY_NAME=GoDownload

# macOS ARM
build-macos-arm:
	GOOS=darwin GOARCH=arm64 go build -o build/$(BINARY_NAME)-macos-arm

# macOS Intel
build-macos-intel:
	GOOS=darwin GOARCH=amd64 go build -o build/$(BINARY_NAME)-macos-intel

# Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o build/$(BINARY_NAME)-linux

# Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o build/$(BINARY_NAME)-windows.exe

# Build all
all: build-macos-arm build-macos-intel build-linux build-windows

# Clean build artifacts
clean:
	rm -rf build/

.PHONY: build-macos-arm build-macos-intel build-linux build-windows all clean
