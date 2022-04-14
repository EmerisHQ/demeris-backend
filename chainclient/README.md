# Chain Client

This chain client will be used for integration testing for multiple chains. There will be configuration files for each chain under test_data/client, our chain client will read data from that folder and return a new client for each chain as required.

#### Reason behind usage of v0.16.1 version of starport client.
```bash
- As demeris backend supports v0.42 version of cosmos sdk, it was required to downgrade the usage of chain client to the v0.16.1 as its supports v0.42 version of [sdk](https://github.com/tendermint/starport/blob/v0.16.1/starport/pkg/spn/spn.go) .

- we have faced import issues while using v0.19.1 [starport client](https://github.com/tendermint/starport/blob/v0.19.1/starport/pkg/cosmosclient/cosmosclient.go) as its supports v0.44 version of sdk.

* we have duplicated most of the code of starport client since its context was not global and we are not able to use it in our code.
```

#### When we have to upgrade our chain client
```bash
- If demeris backend upgrades to v0.44+ version of cosmos sdk, we can updrade our chain client to use latest version of starport. With that most of the duplicated code can be removed and we can directly use starport client methods.
- This an [example](https://github.com/allinbits/demeris-backend/blob/prathyusha/chainclient_v44/chainclient/create_client.go) to upgrade our chain client to use latest version of starport client i.e., v0.19.1.
```

#### How to use chain client
```bash
- In the first step we have to configure chains information under test_data/client/dev or staging.
- This is how config file looks like .
{
    "chain_name":"akash",  ## Give your chain name, for this you can also refer config files under ci/env/chains, env: dev/staging
    "rpc": "https://akash-emeris.app.alpha.starport.cloud:443",  ## RPC of your node
    "key":"akash_test", ## Key name of your account.
    "mnemonic":"Some drink shoot suffer cradle art melt lake crane food cat mask champion force dilemma lizard merit color portion hammer pig portion mix spin",  ##  Mnemonic of the account which you wants to import, if you are testing for staging then you need not to pass mnemonic from here. You can export through env variables. ex : export CHAINNAME_MNEMONIC="seed goes here" (export AKASH_MNEMONIC="seed of your account")
    "keyring_service_name":"akash" ### keyring service name
}

- After configuring chains you can call the method [GetClient](https://github.com/allinbits/demeris-backend/blob/prathyusha/chainclient/chainclient/import_accounts.go#L24) by passing required params. This method will returns a chainclient for that particular chain details you passed to it. This method also imports the mnemonic in keyring and returns the context of it, which you can use to get information about account.
- After getting client you can query the account balances and account information by providing required params.You can find the methods [here](https://github.com/allinbits/demeris-backend/blob/prathyusha/chainclient/chainclient/create_client.go).
- To make a transaction you can call the [broadCastTx](https://github.com/allinbits/demeris-backend/blob/prathyusha/chainclient/chainclient/broadcast_tx.go) method passing params to it.
```
