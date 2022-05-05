package tmwsproxy

import (
	"github.com/emerishq/emeris-utils/configuration"
	"github.com/emerishq/emeris-utils/validation"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	TendermintNode        string `validate:"required"`
	ListenAddr            string `validate:"required"`
	DatabaseConnectionURL string `validate:"required"`
	Debug                 bool
}

func (c *Config) Validate() error {
	err := validator.New().Struct(c)
	if err == nil {
		return nil
	}

	return validation.MissingFieldsErr(err, false)
}

func ReadConfig() (*Config, error) {
	var c Config
	return &c, configuration.ReadConfig(&c, "tmwsproxy", map[string]string{
		"TendermintNode": "http://localhost:26657",
		"ListenAddr":     "localhost:9999",
	})
}
