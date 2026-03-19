PACKNAME=pvtr-gcp-cloud-storage
BUILD_FLAGS=-X 'main.GitCommitHash=`git rev-parse --short HEAD`' -X 'main.BuiltAt=`date +%FT%T%z`'
BUILD_WIN=@env GOOS=windows GOARCH=amd64 go build -o $(PACKNAME).exe
BUILD_LINUX=@env GOOS=linux GOARCH=amd64 go build -o $(PACKNAME)
BUILD_MAC=@env GOOS=darwin GOARCH=amd64 go build -o $(PACKNAME)-darwin

release: package release bin
release-candidate: package release-candidate
binary: package build

release: release-nix release-win release-mac

build:
	@echo "  >  Building binary ..."
	@go build -o $(PACKNAME) -ldflags="$(BUILD_FLAGS)"

package: tidy test
	@echo "  >  Packaging static files..."

test:
	@echo "  >  Validating code ..."
	@go vet ./...
	@go clean -testcache
	@go test ./...

tidy:
	@echo "  >  Tidying go.mod ..."
	@go mod tidy

test-cov:
	@echo "Running tests and generating coverage output ..."
	@go test ./... -coverprofile coverage.out -covermode count
	@sleep 2 # Sleeping to allow for coverage.out file to get generated
	@echo "Current test coverage : $(shell go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+') %"

release-candidate: tidy test
	@echo "  >  Building release candidate for Linux..."
	$(BUILD_LINUX) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=nix-rc'"
	@echo "  >  Building release candidate for Windows..."
	$(BUILD_WIN) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=win-rc'"
	@echo "  >  Building release for Darwin..."
	$(BUILD_MAC) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=darwin-rc'"

release-nix:
	@echo "  >  Building release for Linux..."
	$(BUILD_LINUX) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=linux'"

release-win:
	@echo "  >  Building release for Windows..."
	$(BUILD_WIN) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=windows'"

release-mac:
	@echo "  >  Building release for Darwin..."
	$(BUILD_MAC) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=darwin'"

bin:
	@mv $(PACKNAME)* ~/privateer/bin
