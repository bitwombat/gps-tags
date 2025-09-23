SHELL=/bin/bash

.PHONY: test
test:
	gotestsum ./... -- -test.short

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: test-watch
test-watch:
		find . -name '*.go' -o -name '*.html' | entr sh -c 'gotestsum --max-fails 1 ./... -- -test.short'

.PHONY: test-watch-tmux
test-watch-tmux:
		find . -name '*.go' -o -name '*.html' | entr sh -c 'tmux display-popup -EE gotestsum --max-fails 1 ./... -- -failfast -test.short'

.PHONY: live-test
live-test:
		\go test live_test.go

