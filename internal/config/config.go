package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/url"
	"os"
)

type VaultHysteriaConfig struct {
	VaultToken            string                   `json:"vaultToken"`
	VaultAddr             string                   `json:"vaultAddr"`
	TokenIncreaseSeconds  int                      `json:"tokenIncreaseSeconds"`
	TokenIncreaseInterval int                      `json:"tokenIncreaseInterval"`
	Adapters              []map[string]interface{} `json:"adapters"`
	Filter                map[string]interface{}   `json:"messageFilter"`
}

func (conf *VaultHysteriaConfig) IsTokenIncreaseEnabled() bool {
	return conf.TokenIncreaseInterval > 0 || conf.TokenIncreaseSeconds > 0
}

func AcmeVaultClientConfigFromFile(path string) (VaultHysteriaConfig, error) {
	conf := DefaultVaultConfig()

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, fmt.Errorf("can not read config from file %s: %v", path, err)
	}

	err = json.Unmarshal(content, &conf)
	return conf, err
}

func (conf *VaultHysteriaConfig) Print() {
	log.Info().Msgf("VaultAddr=%s", conf.VaultAddr)
	log.Info().Msgf("VaultToken=%s", "*** (Redacted)")
	if conf.TokenIncreaseSeconds > 0 {
		log.Info().Msgf("TokenIncreaseSeconds=%d", conf.TokenIncreaseSeconds)
	}
	if conf.TokenIncreaseInterval > 0 {
		log.Info().Msgf("TokenIncreaseInterval=%d", conf.TokenIncreaseInterval)
	}
}

func DefaultVaultConfig() VaultHysteriaConfig {
	return VaultHysteriaConfig{
		VaultAddr:  os.Getenv("VAULT_ADDR"),
		VaultToken: os.Getenv("VAULT_TOKEN"),
	}
}

func (conf *VaultHysteriaConfig) Validate() error {
	if len(conf.VaultAddr) == 0 {
		return errors.New("no Vault address defined")
	}
	addr, err := url.ParseRequestURI(conf.VaultAddr)
	if err != nil || addr.Scheme == "" || addr.Host == "" || addr.Port() == "" {
		return errors.New("can not parse supplied vault addr as url")
	}

	if len(conf.VaultToken) == 0 {
		return errors.New("no login token defined")
	}

	return nil
}
