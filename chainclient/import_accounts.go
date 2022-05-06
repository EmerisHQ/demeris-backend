package client

import (
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emerishq/demeris-backend-models/cns"
)

const (
	StagingEnvKey      = "staging"
	AkashMnemonicKey   = "AKASH_MNEMONIC"
	CosmosMnemonicKey  = "COSMOS_MNEMONIC"
	TerraMnemonicKey   = "TERRA_MNEMONIC"
	OsmosisMnemonicKey = "OSMOSIS_MNEMONIC"
)

// GetClient is to create client and imports mnemonic and returns created chain client
func GetClient(env string, chainName string, cc ChainClient, dir string) (c *ChainClient, err error) {
	// get chain info
	info, err := LoadSingleChainInfo(env, chainName)
	if err != nil {
		return nil, err
	}

	initSDKConfig(info.NodeInfo.Bech32Config)
	c, err = CreateChainClient(cc.RPC, cc.KeyringServiceName, info.NodeInfo.ChainID, dir)
	if err != nil {
		return nil, err
	}

	mnemonic := cc.Mnemonic
	if env == StagingEnvKey {
		mnemonic = GetMnemonic(chainName)
	}

	c.AddressPrefix = info.NodeInfo.Bech32Config.PrefixAccount
	c.HDPath = info.DerivationPath
	c.Enabled = info.Enabled
	c.ChainName = info.ChainName
	c.Mnemonic = mnemonic
	c.ChainName = chainName
	if len(info.Denoms) != 0 {
		c.Denom = info.Denoms[0].Name
	}

	_, err = c.ImportMnemonic(cc.Key, c.Mnemonic, c.HDPath)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func initSDKConfig(config cns.Bech32Config) {
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(config.Bech32PrefixAccAddr(), config.Bech32PrefixAccPub())
	sdkConfig.SetBech32PrefixForValidator(config.Bech32PrefixValAddr(), config.Bech32PrefixValPub())
	sdkConfig.SetBech32PrefixForConsensusNode(config.Bech32PrefixConsAddr(), config.Bech32PrefixConsPub())
}

// GetMnemonic returns the mnemonic of particular chain for staging accounts
func GetMnemonic(chName string) string {
	var mnemonic string

	switch chName {
	case "akash":
		mnemonic = os.Getenv(AkashMnemonicKey)
	case "cosmos-hub":
		mnemonic = os.Getenv(CosmosMnemonicKey)
	case "terra":
		mnemonic = os.Getenv(TerraMnemonicKey)
	case "osmosis":
		mnemonic = os.Getenv(OsmosisMnemonicKey)
	default:
		mnemonic = os.Getenv("MNEMONIC")
	}

	return mnemonic
}
