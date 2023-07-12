module github.com/soerenschneider/vault-hysteria

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/hashicorp/go-retryablehttp v0.7.4
	github.com/hashicorp/vault/api v1.9.2
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.42.0
	github.com/rs/zerolog v1.29.1
)

require golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
