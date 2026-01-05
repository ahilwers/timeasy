build:
	@go build -o bin/timeasy-server ./cmd/service/main.go
run: build
	@./bin/timeasy-server
test:
	@go test -p=1 -v ./...
richtest:
	@richgo test -p=1 -v ./...
