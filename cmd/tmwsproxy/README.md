# tmwsproxy

A websocket event proxy for Tendermint RPC.

## Building

```shell
go build -v --ldflags="-s -w"
```

To compile a debug build, remove `ldflags` parameter.

## What does it do?

`tmwsproxy` proxies block and transaction events that happens on a Tendermint instance to another Websocket consumer.

Its initial application is proxying to a IBC relayer only a subset of the transaction events that happen on a given
chain, hence achieveing the objective of only relaying the transactions a relay user is really interested in.

It opens a websocket server on one end, and on the other end it opens a websocket client to the configured Tendermint
RPC.

With a given logic, embedded in `proxy.go` one is able to determine what transaction events to relay to the IBC relayer.

## Configuring tmwsproxy

|Configuration value|Default value|Required|Meaning|
| --- | --- | --- | --- |
|`TendermintNode`|`http://localhost:26657`|no|Original Tendermint RPC to source events from|
|`ListenAddr`|`localhost:9999`|no|Address on which listen for Websocket connections|
|`Debug`|`false`|no|Enable debug logs|

`tmwsproxy` is also configurable through environment variables.

Prepend the string `TMWSPROXY_` to each configuration value, all caps.