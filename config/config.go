package config

import (
	"github.com/allinbits/navigator-utils/validation"

	"github.com/allinbits/navigator-utils/configuration"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	SshUser string
	SshHost string
	SshPort string
	KeyFile string
	DbHost  string
	DbPort  uint16
	DbUser  string
	DbPass  string
	UseSsh  bool
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

	return &c, configuration.ReadConfig(&c, "navigator-api", map[string]string{})
}
