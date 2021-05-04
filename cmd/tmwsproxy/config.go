package main

import (
	"fmt"

	"github.com/allinbits/demeris-backend/utils/configuration"
)

type config struct {
	TendermintNode string
	ListenAddr     string
	Debug          bool
}

func (c *config) Validate() error {
	if c.TendermintNode == "" {
		return fmt.Errorf("missing tendermint node path")
	}

	if c.ListenAddr == "" {
		return fmt.Errorf("missing listen address")
	}

	return nil
}

func readConfig() (*config, error) {
	var c config
	return &c, configuration.ReadConfig(&c, "tmwsproxy", map[string]string{
		"TendermintNode": "http://localhost:26657",
		"ListenAddr":     "localhost:9999",
	})
}
