module github.com/soerenschneider/vault-hysteria

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/hashicorp/go-retryablehttp v0.7.4
	github.com/hashicorp/vault/api v1.10.0
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/common v0.44.0
	github.com/rs/zerolog v1.31.0
)

require golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
