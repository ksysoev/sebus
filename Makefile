test:
	go test -v --race ./...

lint:
	golangci-lint run

bench:
	go test -bench=. -benchmem -benchtime=1s