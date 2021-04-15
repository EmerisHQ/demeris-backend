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
	Title               string        `json:"title"`
	SSLMode             bool          `json:"mode"`
	SSLCrt              string        `json:"sslcrt"`
	SSLKey              string        `json:"sslkey"`
	DB                  string        `json:"db"`
	RestIpPort          string        `json:"rest_ip_port"`
	Interval            time.Duration `json:"interval"`
	WhitelistTokens     []string      `json:"whitelist_tokens"`
	WhitelistFiats      []string      `json:"whitelist_fiats"`
	CoinmarketcapAPIKey string        `json:"coinmarketcap_key"`

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
	var config configType

	path := "config.toml"

	if _, err := toml.DecodeFile(path, &config); err != nil {
		logger.Fatal("Fatal",
			zap.String("Config", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
}
