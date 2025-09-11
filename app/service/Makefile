SHELL=/bin/bash

.PHONY: test
test:
	gotestsum ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: test-watch
test-watch:
		find . -name '*.go' -o -name '*.html' | entr sh -c 'gotestsum --max-fails 1 ./...'

.PHONY: test-watch-tmux
test-watch-tmux:
		find . -name '*.go' -o -name '*.html' | entr sh -c 'tmux display-popup -EE gotestsum --max-fails 1 ./... -- -failfast'
