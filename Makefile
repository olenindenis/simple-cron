build:
	env GOOS=linux GOARCH=amd64 go build -v -o dist/cron cmd/main.go