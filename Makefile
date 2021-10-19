BUILD_DIR = builds
MODULE = github.com/soerenschneider/vault-hysteria
BINARY_NAME = vault-hysteria
CHECKSUM_FILE = $(BUILD_DIR)/checksum.sha256
SIGNATURE_KEYFILE = ~/.signify/github.sec
DOCKER_PREFIX = ghcr.io/soerenschneider

tests:
	go test ./...

clean:
	git diff --quiet || { echo 'Dirty work tree' ; false; }
	rm -rf ./$(BUILD_DIR)

build: version-info
	CGO_ENABLED=0 go build -ldflags="-X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}'" -o $(BINARY_NAME) cmd/main.go

release: clean version-info cross-build-client cross-build-server
	sha256sum $(BUILD_DIR)/acmevault-* > $(CHECKSUM_FILE)

signed-release: release
	pass keys/signify/github | signify -S -s $(SIGNATURE_KEYFILE) -m $(CHECKSUM_FILE)

cross-build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0       go build -ldflags="-X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64     cmd/main.go
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags="-X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv6     cmd/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0       go build -ldflags="-X '$(MODULE)/internal.BuildVersion=${VERSION}' -X '$(MODULE)/internal.CommitHash=${COMMIT_HASH}'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-aarch64   cmd/main.go

version-info:
	$(eval VERSION := $(shell git describe --tags --abbrev=0 || echo "dev"))
	$(eval COMMIT_HASH := $(shell git rev-parse HEAD))

fmt:
	find . -iname "*.go" -exec go fmt {} \; 

pre-commit-init:
	pre-commit install
	pre-commit install --hook-type commit-msg

pre-commit-update:
	pre-commit autoupdate