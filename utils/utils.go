package utils

import (
	"github.com/allinbits/demeris-backend/cns/cnsdb"
)

func GetAddChainParams(chain cnsdb.Chain) cnsdb.AddChainParams {
	return cnsdb.AddChainParams{
		ChainName:        chain.ChainName,
		Enabled:          chain.Enabled,
		Logo:             chain.Logo,
		DisplayName:      chain.DisplayName,
		ValidBlockThresh: chain.ValidBlockThresh,
		PrimaryChannel:   chain.PrimaryChannel,
		Denoms:           chain.Denoms,
		DemerisAddresses: chain.DemerisAddresses,
		GenesisHash:      chain.GenesisHash,
		NodeInfo:         chain.NodeInfo,
		DerivationPath:   chain.DerivationPath,
		BlockExplorer:    chain.BlockExplorer,
	}
}
