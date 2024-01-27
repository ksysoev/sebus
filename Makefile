test:
	docker-compose run --rm tests

lint:
	golangci-lint run