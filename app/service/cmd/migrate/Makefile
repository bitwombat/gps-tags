
dogs_pre_20250711.sql:
	-@[ -e dogtags.db ] && rm -i dogtags.db
	go run main.go dogs_pre_20250711.json

.PHONY: data
data:
	@if ! `which sqlite3 >/dev/null 2>&1`; then echo "Can't find sqlite3 executable"; else \
		( echo "importing old data (slow)" ; cat dogs_pre_20250711.sql | sqlite3 -bail dogtags.db ) ;\
	fi

.PHONY: all
all: dogs_pre_20250711.sql data

