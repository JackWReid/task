.PHONY: build install release test clean version

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS = -X github.com/jackreid/task/internal/version.Version=$(VERSION) \
          -X github.com/jackreid/task/internal/version.BuildDate=$(BUILD_DATE) \
          -X github.com/jackreid/task/internal/version.GitCommit=$(GIT_COMMIT)

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o task .

# Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./...

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f task task.exe *.test

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"
