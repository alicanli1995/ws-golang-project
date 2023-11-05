#!/bin/zsh

go build -o vigilate cmd/web/*.go && ./vigilate \
-db='postgres' \
-dbuser='postgres' \
-dbpass='postgres' \
-pusherHost='localhost' \
-pusherKey='abc123' \
-pusherSecret='123abc' \
-pusherApp="1" \
-pusherPort="4001" \
-pusherSecure=false