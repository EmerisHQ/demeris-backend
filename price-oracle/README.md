## Usage

### Configuration
Description of the `demeris-price-oracle.toml` setting file.

Key Feature Settings

- interval: The delay time of the function for an external price provider api request.
- whitelistfiats : List of fiats you want to request (default currency is USD)
- coinmarketcapapikey : This is the api-key of the provider.
- fixerapikey : This is the api-key of the provider.
- Provider : The endpoint address of the price provider.

For Binance, apikey does not exist.

example(Local exec)
```bash
#kubectl port-forward service/cockroachdb-public 26257
DatabaseConnectionURL = "postgres://root@127.0.0.1:26257?sslmode=disable"
ListenAddr = "127.0.0.1:9898"
Debug = true
LogPath = "/home/ubuntu/log"
interval = "10s"
whitelistfiats = ["EUR","KRW","CHF"]
#Not currently used, but may be used in the future
#coinmarketcapapikey =""
fixerapikey = ""
```

### Local exec DB
`database/schema`
Set the cockroach DB to the local cluster, connect to the local DB, and run schema as it is.

*The cns version brunch does not require a separate run.

### Build

```bash
# build
go build

# executable
./navigator-price-oracle
```

### Use
Given an exchange API, the price oracle will periodically check with it prices of the tokens we're interested in monitoring, and cache the result until the next period:

1. at startup, oracle will grab data from the API and expose it on `GET /prices`
2. once every 10 seconds, new data will be downloaded and cached

Oracle must return prices of all the tokens that it is configured to fetch.

An API to provide tokens configuration must be provided on `POST /tokens`:

```jsx
{
	"tokens": [
		"ATOMUSDT",
		"KAVAUSDT"
	]
}
```

The same must be done for FIAT currencies on `POST /fiats`:

```jsx
{
	"fiats": [
		"USDEUR",
		"USDKRW"
	]
}
```