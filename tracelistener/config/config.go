package config

import (
	"github.com/allinbits/demeris-backend/utils/validation"

	"github.com/allinbits/demeris-backend/utils/configuration"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	FIFOPath              string `validate:"required"`
	ChainName             string `validate:"required"`
	DatabaseConnectionURL string `validate:"required"`
	LogPath               string
	SDKVersion            string `validate:"required"`
	Debug                 bool

	// Processors configs
	ProcessorConfig ProcessorConfig
}

type ProcessorConfig struct {
	ProcessorsEnabled []string
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}

	return validation.MissingFieldsErr(err, false)
}

func Read() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "tracelistener", map[string]string{
		"FIFOPath": "./.tracelistener.fifo",
	})
}
