.PHONY: build
build: deps
	go build

.PHONY: deps
deps:
	go get -v -t

.PHONY: test
test: deps
	go test -v
