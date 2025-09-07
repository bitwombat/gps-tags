SHELL=/bin/bash
.PHONY: test

test:
	gotestsum ./...

test-watch:
		find . -name '*.go' -o -name '*.html' | entr sh -c 'gotestsum ./...'

test-watch-tmux:
		find . -name '*.go' -o -name '*.html' | entr sh -c 'tmux display-popup -EE gotestsum ./...'
