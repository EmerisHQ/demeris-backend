package server

import "time"

type Prices struct {
	Symbol string `db:"symbol"`
	Prirce string `db:"price"`
	//time.now add insert
}

type SelectToken struct {
	Tokens []string `json:"tokens"`
}

type SelectFiat struct {
	Fiat []string `json:"fiat"`
}

type Binance struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Coinmarketcap struct {
	Status struct {
		Timestamp    time.Time   `json:"timestamp"`
		ErrorCode    int         `json:"error_code"`
		ErrorMessage interface{} `json:"error_message"`
		Elapsed      int         `json:"elapsed"`
		CreditCount  int         `json:"credit_count"`
		Notice       interface{} `json:"notice"`
	} `json:"status"`
	Data struct {
		Atom struct {
			ID                int         `json:"id"`
			Name              string      `json:"name"`
			Symbol            string      `json:"symbol"`
			Slug              string      `json:"slug"`
			NumMarketPairs    int         `json:"num_market_pairs"`
			DateAdded         time.Time   `json:"date_added"`
			Tags              []string    `json:"tags"`
			MaxSupply         interface{} `json:"max_supply"`
			CirculatingSupply float64     `json:"circulating_supply"`
			TotalSupply       float64     `json:"total_supply"`
			IsActive          int         `json:"is_active"`
			Platform          interface{} `json:"platform"`
			CmcRank           int         `json:"cmc_rank"`
			IsFiat            int         `json:"is_fiat"`
			LastUpdated       time.Time   `json:"last_updated"`
			Quote             struct {
				Usdt struct {
					Price            float64   `json:"price"`
					Volume24H        float64   `json:"volume_24h"`
					PercentChange1H  float64   `json:"percent_change_1h"`
					PercentChange24H float64   `json:"percent_change_24h"`
					PercentChange7D  float64   `json:"percent_change_7d"`
					PercentChange30D float64   `json:"percent_change_30d"`
					PercentChange60D float64   `json:"percent_change_60d"`
					PercentChange90D float64   `json:"percent_change_90d"`
					MarketCap        float64   `json:"market_cap"`
					LastUpdated      time.Time `json:"last_updated"`
				} `json:"USDT"`
			} `json:"quote"`
		} `json:"ATOM"`
	} `json:"data"`
}
