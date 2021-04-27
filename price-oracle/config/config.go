package config

import (
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

var (
	Config configType
)

type configType struct {
	SSLMode             bool          `json:"sslmode"`
	SSLCrt              string        `json:"sslcrt"`
	SSLKey              string        `json:"sslkey"`
	DB                  string        `json:"db"`
	Laddr               string        `json:"laddr"`
	Interval            time.Duration `json:"interval"`
	Whitelisttokens     []string      `json:"whitelisttokens"`
	Whitelistfiats      []string      `json:"whitelistfiats"`
	CoinmarketcapapiKey string        `json:"coinmarketcapapikey"`

	APIs struct {
		Atom struct {
			Usd struct {
				Binance       string `json:"binance"`
				Coinmarketcap string `json:"coinmarketcap"`
			}
		}

		Stables struct {
			Currencylayer string `json:"currencylayer"`
		}

		Sdr struct {
			Imf string `json:"imf"`
		}

		Band struct {
			Active bool   `json:"active"`
			Band   string `json:"band"`
		}
	}
}

func ReadConfig(logger *zap.Logger) {
	if _, err := toml.DecodeFile("config.toml", &Config); err != nil {
		logger.Fatal("Fatal",
			zap.String("Config", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
}
