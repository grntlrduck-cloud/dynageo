
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

test_report:
	go run --mod=mod gotest.tools/gotestsum --junitfile unit-tests.xml  -- -coverprofile=cover.out -covermode count -p 1 ./...
	go tool cover -html=cover.out -o coverage.html
	go run --mod=mod github.com/boumenot/gocover-cobertura <cover.out > coverage.xml
