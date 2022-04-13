compileproto:
	protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
./proto/2dfortnite.proto

build_client:
	go build -o bin/2dfortnite ./src/client

build_server:
	go build -o bin/2dfortnite_server ./src/server/

build: build_client build_server