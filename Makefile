build:
	env GOOS=linux GOARCH=amd64 go build -v -o dist/cron cmd/main.go

# Native arm64 build for arm64 Docker/VM hosts (e.g. Apple Silicon) to avoid
# running the amd64 binary under Rosetta/QEMU emulation.
build-arm64:
	env GOOS=linux GOARCH=arm64 go build -v -o dist/cron-arm64 cmd/main.go

test:
	go test ./... -race -v

lint:
	docker run -t --rm -v $$(pwd):/app -w /app golangci/golangci-lint:v2.12.2 golangci-lint run