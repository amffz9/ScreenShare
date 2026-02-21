APP_NAME := ScreenShare
LDFLAGS := -ldflags="-s -w"
DIST := dist

.PHONY: build build-all clean run icon

# Generate Windows icon/manifest resource (requires: go install github.com/tc-hib/go-winres@latest)
icon:
	go-winres make --arch amd64

build:
	go build $(LDFLAGS) -o $(DIST)/$(APP_NAME) .

build-all:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP_NAME)-Windows.exe .
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP_NAME)-Linux .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP_NAME)-Mac-Intel .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(APP_NAME)-Mac-AppleSilicon . || echo "NOTE: darwin/arm64 requires Go 1.21+, skipping"

clean:
	rm -rf $(DIST)

run:
	go run .
