#!/bin/bash

set -e

DBNAME="dogtags.db"
JSON="dogs_pre_20250711.json"
SQL=$(basename "$JSON" .json).sql

if [ ! -x "$(which sqlite3 2>/dev/null)" ]; then
    echo "Can't find sqlite3 executable"
    exit 1
fi

[ -e "$DBNAME" ] && rm -i "$DBNAME"

go run . "$JSON"

echo "importing old data (slow)"
cat "$SQL" | sqlite3 -bail "$DBNAME"
