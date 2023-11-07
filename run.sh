#!/bin/zsh

go build -o observer cmd/web/*.go && ./observer \
-db='postgres' \
-dbuser='postgres' \
-dbpass='postgres' \
-pusherHost='localhost' \
-pusherKey='abc123' \
-pusherSecret='123abc' \
-pusherApp="1" \
-pusherPort="4001" \
-pusherSecure=false