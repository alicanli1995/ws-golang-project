DB=postgres
DB_USER=postgres
DB_PASSWORD=postgres
PUSHER_HOST=localhost
PUSHER_KEY=abc123
PUSHER_SECRET=123abc
PUSHER_APP=1
PUSHER_PORT=4001
PUSHER_SECURE=false

## start_ipe: starts the IPE
start_ipe:
	@echo Starting the IPE...
	@cd ./ipe && ipe.exe
	@echo IPE running!

## build_back: builds the back end
build_back:
	@echo Building back end...
	@go build -o observer.exe ./cmd/web/
	@echo Back end built!

## start_back: starts the back end
start_back: build_back
	@echo Starting the back end...
	go build -o observer.exe .\cmd\web\ &&observer.exe -dbuser=${DB_USER} -pusherHost=${PUSHER_HOST} \
	-pusherSecret=${PUSHER_SECRET} -pusherKey=${PUSHER_KEY} -pusherSecure=${PUSHER_SECURE} -pusherApp=${PUSHER_APP} -db=${DB} -dbpass=${DB_PASSWORD}
	@echo Back end running!
