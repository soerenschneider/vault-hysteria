package main

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/vault-hysteria/internal"
	"github.com/soerenschneider/vault-hysteria/internal/adapters"
	httpAdapter "github.com/soerenschneider/vault-hysteria/internal/adapters/http"
	"github.com/soerenschneider/vault-hysteria/internal/config"
	"github.com/soerenschneider/vault-hysteria/internal/messagefilters"
	anyFilter "github.com/soerenschneider/vault-hysteria/internal/messagefilters/any"
	containsFilter "github.com/soerenschneider/vault-hysteria/internal/messagefilters/contains"
)

func main() {
	// TODO: Make configurable
	conf, err := config.AcmeVaultClientConfigFromFile("conf.json")
	if err != nil {
		log.Fatal().Msgf("Could not load config: %v", err)
	}
	conf.Print()
	err = conf.Validate()
	if err != nil {
		log.Fatal().Msgf("could not validate config: %v", err)
	}
	vault, err := internal.NewVaultBackend(conf)
	if err != nil {
		log.Fatal().Msgf("could not build vault client: %v", err)
	}

	log.Info().Msg("Building adapters...")
	adapters, err := buildAdaptersFromConfig(conf)
	if err != nil {
		log.Fatal().Msgf("could not build all adapters: %v", err)
	}
	// TODO: Make configurable
	go internal.StartMetricsServer(":9191")

	filter, err := buildFilterFromConfig(conf)
	if err != nil {
		log.Fatal().Msgf("could not build desired filter: %v", err)
	}

	vaultHysteria, err := internal.NewVaultHysteria(vault, adapters, filter)
	if err != nil {
		log.Fatal().Msgf("could not build seal: %v", err)
	}

	vaultHysteria.Start()
}

func buildFilterFromConfig(conf config.VaultHysteriaConfig) (messagefilters.MessageFilter, error) {
	for _, keyword := range []string{"type"} {
		_, ok := conf.Filter[keyword]
		if !ok {
			return nil, fmt.Errorf("no '%s' defined for filter", keyword)
		}
	}
	filterType := conf.Filter["type"]

	var argsMap map[string]interface{}
	_, ok := conf.Filter["args"]
	if ok {
		argsMap, ok = conf.Filter["args"].(map[string]interface{})
		if !ok {
			return nil, errors.New("error parsing args for filter: not a map")
		}
	}
	switch filterType {
	case anyFilter.FilterName:
		return &anyFilter.AnyFilter{}, nil

	case containsFilter.FilterName:
		return containsFilter.NewContainsFilterFromMap(argsMap)
	}

	return nil, fmt.Errorf("no such filter: %s", filterType)
}

func buildAdaptersFromConfig(conf config.VaultHysteriaConfig) ([]adapters.Adapter, error) {
	builtAdapters := make([]adapters.Adapter, 0)

	for _, adapter := range conf.Adapters {
		for _, keyword := range []string{"type", "args"} {
			_, ok := conf.Filter[keyword]
			if !ok {
				return nil, fmt.Errorf("no '%s' defined for filter", keyword)
			}
		}

		adapterType, _ := adapter["type"]
		argsMap, ok := adapter["args"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error parsing args for adapter %s: not a map", adapterType)
		}

		if adapterType == httpAdapter.AdapterName {
			httpListener, err := httpAdapter.HttpListenerFromConfigMap(argsMap)
			if err != nil {
				return nil, fmt.Errorf("could not build requested http listener: %v", err)
			}
			builtAdapters = append(builtAdapters, httpListener)
		} else {
			return nil, fmt.Errorf("don't know how to build adapter '%s'", adapterType)
		}
	}

	return builtAdapters, nil
}
