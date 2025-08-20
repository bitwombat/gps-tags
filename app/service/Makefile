SHELL=/bin/bash
.PHONY: test

test:
	-@docker stop mongodb-test &> /dev/null
	-@docker rm mongodb-test &> /dev/null
	echo "Starting mongodb" && docker run --name mongodb-test --quiet --rm --detach -p 27017:27017 mongodb/mongodb-community-server:8.0-ubi9
	#go test ./...
	gotestsum --format testname --post-run-command beep -- ./...
	echo "Stopping monogdb" && docker stop mongodb-test &> /dev/null

test-watch:
	-@docker stop mongodb-test &> /dev/null
	-@docker rm mongodb-test &> /dev/null
	echo "Starting mongodb" && docker run --name mongodb-test --quiet --rm --detach -p 27017:27017 mongodb/mongodb-community-server:8.0-ubi9
	#go test ./...
	gotestsum --format testname --watch --post-run-command beep -- ./...
	echo "Stopping monogdb" && docker stop mongodb-test &> /dev/null
