package main

import (
	"fmt"

	"github.com/allinbits/demeris-backend/utils/configuration"
	"github.com/allinbits/demeris-backend/utils/validation"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Redis               string `validate:"required"`
	KubernetesNamespace string `validate:"required"`
	LogPath             string
	Debug               bool
}

func (c Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}
	return fmt.Errorf(
		"configuration file error: %w",
		validation.MissingFieldsErr(err, false),
	)
}

func readConfig() (*Config, error) {
	var c Config

	return &c, configuration.ReadConfig(&c, "emeris-tw", map[string]string{
		"RedisURL":            "http://localhost:6379",
		"KubernetesNamespace": "emeris",
	})
}
