# Makefile

mockgen:
	go install github.com/golang/mock/mockgen@v1.6.0
	mockgen -source=internal/repository/public.go -destination=internal/repository/public_mock.gen.go -package=repository

lint:
	golangci-lint run

test:
	go test ./...

pre-commit:
	make mockgen
	make lint
	make test
