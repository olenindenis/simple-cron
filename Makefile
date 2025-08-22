build:
	env GOOS=linux GOARCH=amd64 go build -v -o dist/cron_amd64 cmd/main.go