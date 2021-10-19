package internal

import (
	"errors"
	backoff "github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/vault-hysteria/internal/adapters"
	"github.com/soerenschneider/vault-hysteria/internal/messagefilters"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// maxBackoffDuration defines how long to try to seal the vault
const maxBackoffDuration = 1 * time.Hour

// heartbeatFreq defines the frequency to increase the heartbeat metric and to check the vault token
const heartbeatFreq = 60 * time.Second

type Vault interface {
	Seal() error
	Heartbeat()
}

type VaultHysteria struct {
	heartbeat    *time.Ticker
	adapters     []adapters.Adapter
	panicChannel chan string
	accept       messagefilters.MessageFilter
	vault        Vault
}

func NewVaultHysteria(vault Vault, listeners []adapters.Adapter, accept messagefilters.MessageFilter) (*VaultHysteria, error) {
	if listeners == nil || len(listeners) == 0 {
		return nil, errors.New("no adapters provided")
	}

	if vault == nil {
		return nil, errors.New("empty implementation for vault provided")
	}

	return &VaultHysteria{
		heartbeat:    time.NewTicker(heartbeatFreq),
		adapters:     listeners,
		panicChannel: make(chan string),
		accept:       accept,
		vault:        vault,
	}, nil
}

func (v *VaultHysteria) Start() {
	for _, listener := range v.adapters {
		listener.Start(v.panicChannel)
	}

	go func() {
		for {
			<-v.heartbeat.C
			HeartbeatMetric.SetToCurrentTime()
			v.vault.Heartbeat()
		}
	}()

	go v.readFromPanicChannel()

	// watch for signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT)
	<-sig

	v.shutdown()
	log.Info().Msg("Bye")
}

func (v *VaultHysteria) readFromPanicChannel() {
	for read := range v.panicChannel {
		SealRequestsReceived.Inc()
		if v.accept.Accept(read) {
			err := v.seal()
			if err != nil {
				log.Error().Msgf("Gave up trying to v vault: %v", err)
			}
		} else {
			SealRequestsIgnored.Inc()
		}
	}
}

func (v *VaultHysteria) seal() error {
	log.Info().Msg("Sealing vault")
	operation := func() error {
		return v.vault.Seal()
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = time.Duration(maxBackoffDuration)

	err := backoff.RetryNotify(operation, expBackoff, func(err error, duration time.Duration) {
		log.Error().Msgf("Error sealing vault, trying again in %v: %v", duration, err)
		FailedSealCallsMetric.Inc()
	})

	return err
}

func (v *VaultHysteria) shutdown() {
	log.Info().Msg("Closing panic channel..")
	close(v.panicChannel)

	log.Info().Msg("Stopping heartbeat..")
	v.heartbeat.Stop()

	for _, adapter := range v.adapters {
		log.Info().Msgf("Stopping adapter %s", adapter.Name())
		err := adapter.Stop()
		if err != nil {
			log.Error().Msgf("Error stopping adapter %s: %v", adapter.Name(), err)
		}
	}
}
