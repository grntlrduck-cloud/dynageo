
ci:
	go mod download && go mod verify

vuln_scan:
	go run --mod=mod golang.org/x/vuln/cmd/govulncheck ./...

lint:
	golangci-lint run ./...

update:
	go get -u ./...
	go mod tidy
	go mod verify
