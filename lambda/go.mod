module ledger-lambda

go 1.23.0

toolchain go1.23.4

replace Ledger => ../

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.24.0
)
