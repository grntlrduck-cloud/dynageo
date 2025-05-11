module github.com/grntlrduck-cloud/dynageo

go 1.24.2

toolchain go1.25.2

require (
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.14
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.51.0
	github.com/golang/geo v0.0.0-20251008113719-ccf458c07da6
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.31.0 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
