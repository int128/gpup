TARGET := gpup
CIRCLE_TAG ?= HEAD
LDFLAGS := -X main.version=$(CIRCLE_TAG)

all: $(TARGET)

.PHONY: check
check:
	golangci-lint run
	go test -v -race -cover -coverprofile=coverage.out ./...

$(TARGET): $(wildcard *.go)
	go build -o $@ -ldflags "$(LDFLAGS)"

dist:
	# make the zip files for GitHub Releases
	VERSION=$(CIRCLE_TAG) CGO_ENABLED=0 goxzst -d dist/ -i "LICENSE" -o "$(TARGET)" -t "gpup.rb" -- -ldflags "$(LDFLAGS)"

.PHONY: release
release: dist
	# publish to the GitHub Releases
	ghr -u "$(CIRCLE_PROJECT_USERNAME)" -r "$(CIRCLE_PROJECT_REPONAME)" "$(CIRCLE_TAG)" dist/
	# publish to the Homebrew tap repository
	ghcp commit -u "$(CIRCLE_PROJECT_USERNAME)" -r "homebrew-$(CIRCLE_PROJECT_REPONAME)" -m "$(CIRCLE_TAG)" -C dist/ gpup.rb

.PHONY: clean
clean:
	-rm $(TARGET)
	-rm -r dist/
