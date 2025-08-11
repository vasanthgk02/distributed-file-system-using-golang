build:
	@go build -o bin/fs

bootstrap1: build
	@./bin/fs --mode=bootstrap --port=:3000

bootstrap2: build
	@./bin/fs --mode=bootstrap --port=:3001

node: build
	@./bin/fs --mode=node --port=:3002 --bootstrap=:3000,:3001

test:
	@go test -v ./...

clean:
	@rm -rf bin/ *_network/
