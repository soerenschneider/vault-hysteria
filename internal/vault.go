package internal

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/vault-hysteria/internal/config"
	"strconv"
	"time"
)

const (
	timeout        = 3 * time.Second
	backoffRetries = 5
)

type VaultBackend struct {
	client *api.Client
	conf   config.VaultHysteriaConfig
}

func (b *VaultBackend) Seal() error {
	_, err := b.client.Logical().Write("/sys/seal", map[string]interface{}{})
	return err
}

func (b *VaultBackend) Heartbeat() {
	expiry, err := b.GetExpireTime()
	if err != nil {
		log.Error().Msgf("Error looking up self: %v", err)
		VaultCommunicationErrors.Inc()
	} else {
		log.Debug().Msgf("Vault heartbeatFreq received token expiry date %v", expiry)
		VaultTokenExpiryTime.Set(float64(expiry.Unix()))
	}
}

func (b *VaultBackend) GetExpireTime() (*time.Time, error) {
	secret, err := b.client.Auth().Token().LookupSelf()
	if err != nil {
		return nil, fmt.Errorf("could not get expiry date of token: %v", err)
	}
	ttlData, ok := secret.Data["ttl"]
	if !ok {
		return nil, fmt.Errorf("no 'ttl' field found in data portion")
	}
	ttl, err := strconv.Atoi(fmt.Sprint(ttlData))
	if err != nil {
		return nil, fmt.Errorf("expected an int, but couldn't parse 'ttl' value of '%s': %v", ttlData, err)
	}

	expiry := time.Now().Add(time.Duration(ttl) * time.Second)
	return &expiry, nil
}

func (b *VaultBackend) RenewToken(seconds int) {

}

func NewVaultBackend(vaultConfig config.VaultHysteriaConfig) (*VaultBackend, error) {
	config := &api.Config{
		Timeout:    timeout,
		MaxRetries: backoffRetries,
		Backoff:    retryablehttp.DefaultBackoff,
		Address:    vaultConfig.VaultAddr,
	}

	var err error
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("couldn't build client: %v", err)
	}

	// set initial token, can be empty as well, ignore potential errors
	client.SetToken(vaultConfig.VaultToken)
	vault := &VaultBackend{
		client: client,
		conf:   vaultConfig,
	}

	if vault.conf.IsTokenIncreaseEnabled() {
		go func() {
			ticker := time.NewTicker(time.Duration(vault.conf.TokenIncreaseInterval) * time.Second)
			for {
				select {
				case <-ticker.C:
					vault.RenewToken(vault.conf.TokenIncreaseSeconds)
				}
			}
		}()
	}

	return vault, nil
}
