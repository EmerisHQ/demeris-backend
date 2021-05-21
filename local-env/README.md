# `local-env.sh` tip and tricks

 - only use `-b` flags when you *really* need to rebuild containers (e.g., you're working on backend code and need to test your changes)
 - run the script with `-nc` to not automatically deploy chains, this is particularly useful if you're working on the CNS UI
 - you don't need to run `down` and then `up` again every time, if you modified something just run `up` (optionally with `-b`) and Kubernetes will take care of the rest
 - conversely, if you see the cluster behaving strangely tear it down and up again to see if it works, if it doesn't work please send a message on #demeris-implementation

## Useful Kubernetes commands

```shell
kubectl get pods -A # returns a list of all the pods running in your local cluster
kubectl get nodesets # returns a list of all the chain nodes running on your local cluster
kubectl get faucets # returns a list of all the faucets running on your local cluster
kubectl describe {nodeset, faucet} {itemName} # gives an accurate description of the status, template of a nodeset or faucet identified by `itemName'
```

## Handling chains with CNS

To deploy a Cosmos Hub node, `POST localhost:9999/add` this JSON:

```json
{
    "chain_name":"cosmos-hub",
    "logo": "logo url",
    "display_name": "Cosmos Hub",
    "counterparty_names":{
        "cn1": "cn1",
        "cn2": "cn2"
    },
    "primary_channel":{
        "cn1": "cn1",
        "cn2": "cn2"
    },
    "demeris_addresses": ["feeaddress"],
    "denoms": [
        {
            "display_name": "STAKE",
            "name": "stake",
            "verified": true,
            "precision": 6
        },
                {
                    "display_name": "UATOM",
            "name": "uatom",
            "verified": true,
            "precision": 6
        }
    ],
    "base_ibc_fee":1,
    "genesis_hash":"genesis_hash",
    "node_info": {
        "endpoint": "endpoint",
        "chain_id": "chainid",
        "bech32_config": {
            "main_prefix": "main_prefix",
            "prefix_account": "prefix_account",
            "prefix_validator": "prefix_validator",
            "prefix_consensus": "prefix_consensus",
            "prefix_public": "prefix_public",
            "prefix_operator": "prefix_operator"
        }
    },
    "base_tx_fee": {
        "low": 1,
        "average": 22,
        "high": 42
    },
    "node_config": {
        "name": "cosmos-hub",
        "cli_name": "gaiad",
        "testnet_config": {
            "chainID": "demeris-test",
            "stake_amount": "10000000000stake",
            "assets": [
                "100000000000000000stake",
                "100000000000000000uatom"
            ],
            "faucet": {
                "funds": "100000000000000000stake",
                "denom": [
                    "stake"
                ]
            }
        },
        "docker_image": "gcr.io/tendermint-dev/gaia",
        "docker_image_version": "v4.2.1"
    }
}
```

This JSON will deploy a chain named "cosmos-hub" with a faucet containing `100000000000000000stake` tokens.

---

To deploy an Akash node, `POST localhost:9999/add` this JSON:

```json
{
    "chain_name":"akash",
    "logo": "logo url",
    "display_name": "Akash",
    "counterparty_names":{
        "cn1": "cn1",
        "cn2": "cn2"
    },
    "primary_channel":{
        "cn1": "cn1",
        "cn2": "cn2"
    },
    "demeris_addresses": ["feeaddress"],
    "denoms": [
        {
            "display_name": "STAKE",
            "name": "stake",
            "verified": true,
            "precision": 6
        },
                {
                    "display_name": "UACK",
            "name": "uakt",
            "verified": true,
            "precision": 6
        }
    ],
    "base_ibc_fee":1,
    "genesis_hash":"genesis_hash",
    "node_info": {
        "endpoint": "endpoint",
        "chain_id": "chainid",
        "bech32_config": {
            "main_prefix": "main_prefix",
            "prefix_account": "prefix_account",
            "prefix_validator": "prefix_validator",
            "prefix_consensus": "prefix_consensus",
            "prefix_public": "prefix_public",
            "prefix_operator": "prefix_operator"
        }
    },
    "base_tx_fee": {
        "low": 1,
        "average": 22,
        "high": 42
    },
    "node_config": {
        "name": "akash",
        "testnet_config": {
            "chainID": "demeris-test",
            "stake_amount": "10000000000stake",
            "assets": [
                "100000000000000000stake",
                "100000000000000000uakt"
            ],
           "faucet": {
             "funds": "100000000000000000stake",
             "denom": [
               "stake"
             ]
           }          
        },
        "docker_image": "gcr.io/tendermint-dev/akash",
        "docker_image_version": "v0.12.1"
    }
}
```

## Deleting chains

To delete a chain, `DELETE localhost:9999/delete` the following JSON

```json
{
  "chain": "your chain name"
}
```
