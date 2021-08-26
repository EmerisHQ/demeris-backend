package sdk

import (
	"fmt"

	authV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/auth/v042"
	bankV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/bank/v042"
	bankV043 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/bank/v043"
	delegationsV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/delegations/v042"
	ibcChannelsV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/ibc_channels/v042"
	ibcClientsV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/ibc_clients/v042"
	ibcConnectionsV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/ibc_connections/v042"
	ibcDenomTracesV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/ibc_denom_traces/v042"
	liquidityPoolV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/liquidity_pool/v042"
	liquiditySwapsV042 "github.com/allinbits/demeris-backend/tracelistener/processor/sdk/liquidity_swaps/v042"
)

type PresetList map[string]Preset

type Preset map[string]interface{}

func (p Preset) Module(name string) (interface{}, error) {
	m, ok := p[name]
	if !ok {
		return nil, fmt.Errorf("parser for version %v not found", name)
	}

	return m, nil
}

var v042Preset = Preset{
	"bank":             bankV042.Bank{},
	"auth":             authV042.Auth{},
	"delegations":      delegationsV042.Delegations{},
	"ibc_channels":     ibcChannelsV042.IBCChannels{},
	"ibc_clients":      ibcClientsV042.IBCClients{},
	"ibc_connections":  ibcConnectionsV042.IBCConnections{},
	"ibc_denom_traces": ibcDenomTracesV042.IBCDenomTraces{},
	"liquidity_pool":   liquidityPoolV042.LiquidityPool{},
	"liquidity_swaps":  liquiditySwapsV042.LiquiditySwaps{},
}

var v043Preset = Preset{
	"bank":             bankV043.Bank{},
	"auth":             authV042.Auth{},
	"delegations":      delegationsV042.Delegations{},
	"ibc_channels":     ibcChannelsV042.IBCChannels{},
	"ibc_clients":      ibcClientsV042.IBCClients{},
	"ibc_connections":  ibcConnectionsV042.IBCConnections{},
	"ibc_denom_traces": ibcDenomTracesV042.IBCDenomTraces{},
	"liquidity_pool":   liquidityPoolV042.LiquidityPool{},
	"liquidity_swaps":  liquiditySwapsV042.LiquiditySwaps{},
}

var presets = PresetList{
	"v042": v042Preset,
	"v043": v043Preset,
}

func GetPreset(version string) (Preset, error) {
	p, ok := presets[version]
	if !ok {
		return nil, fmt.Errorf("preset for version %v not found", version)
	}

	return p, nil
}
