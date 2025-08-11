build:
	@go build -o bin/fs

bootstrap1: build
	@echo "Starting bootstrap node on port 3000..."
	@./bin/fs --mode=bootstrap --port=:3000

bootstrap2: build
	@echo "Starting bootstrap node on port 3001..."
	@./bin/fs --mode=bootstrap --port=:3001

node: build
	@echo "Starting node on port 3001 connected to bootstrap..."
	@./bin/fs --mode=node --port=:3002 --bootstrap=:3000,:3001

test:
	@go test -v ./...

clean:
	@rm -rf bin/ *_network/
