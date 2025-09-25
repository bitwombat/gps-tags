SHELL=/bin/bash
PROJECT_FILES=$(shell find . ../public_html -name '*.go' -o -wholename '*public_html/*')

.PHONY: test
test:
	gotestsum --max-fails 1 -- ./... -test.short -failfast

.PHONY: native-test
native-test:
	\go test ./... -v -test.short -failfast

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: test-watch
test-watch:
		echo "${PROJECT_FILES}" | xargs -n 1 | entr -c sh -c 'make --no-print-directory test'

.PHONY: test-watch-tmux
test-watch-tmux:
		echo "${PROJECT_FILES}" | xargs -n 1 | entr sh -c 'tmux display-popup -EE make --no-print-directory test'

.PHONY: live-test
live-test:
		\go test live_test.go

