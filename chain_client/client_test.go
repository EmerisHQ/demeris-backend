package client

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

const (
	Bech32PrefixAccAddr = "akash"
	Bech32PrefixAccPub  = "akashpub"

	Bech32PrefixValAddr = "akashvaloper"
	Bech32PrefixValPub  = "akashvaloperpub"

	Bech32PrefixConsAddr = "akashvalcons"
	Bech32PrefixConsPub  = "akashvalconspub"
)

func TestClient(t *testing.T) {
	mnemonics := []string{
		"aerobic swim trophy document boost flee depend hope cactus science gossip bike tree marine congress elbow ring error spare sniff quiz interest diagram cube",
		"obey exclude rifle evil credit kidney basic elephant light custom happy muscle radio old retire noble body trap kit final song option crumble syrup",
	}

	address := []string{
		"cosmos1c3yme7jsgj83p92fcqwns9zp3cssnerf3k3q4k",
		"akash1kzpa4zlvxyqlz4kqe0rneap3l5c9r2x3u3cse0",
	}

	prefixes := []string{
		"cosmos",
		"akash",
	}

	rpcs := []string{
		"http://localhost:26657",
		"http://localhost:27657",
	}

	for i := 0; i < 2; i++ {
		testChain(t, mnemonics[i], rpcs[i], prefixes[i], address[i])
	}
	require.False(t, true)
}

func testChain(t *testing.T, mnemonic, rpc, prefix, addr string) {
	dir := t.TempDir()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(prefix, prefix+"pub")
	config.SetBech32PrefixForValidator(prefix+"valoper", prefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(prefix+"valcons", prefix+"valconspub")
	// config.Seal()

	client, err := CreateChainClient(rpc, "test-client", "testchain", dir)
	require.NoError(t, err)

	acc, err := client.ImportMnemonic("testAcc", mnemonic, "m/44'/118'/0'/0/0")
	require.NoError(t, err)
	coins, err := client.GetAccountBalances(acc.Address, "stake")
	require.NoError(t, err)
	t.Log("Out...", coins)

	addr1, err := sdk.AccAddressFromBech32(acc.Address)
	require.NoError(t, err)

	addr2, err := sdk.AccAddressFromBech32(addr)
	require.NoError(t, err)

	msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10000))))
	t.Log("Msg...", msg.String())

	out, err := client.Broadcast("testAcc", context.Background(), client.clientCtx, msg)
	require.NoError(t, err)

	t.Log("Hash...", out.TxHash)

	coins, err = client.GetAccountBalances(acc.Address, "stake")
	require.NoError(t, err)
	t.Log("Out 1...", coins)

	coins, err = client.GetAccountBalances(addr, "stake")
	require.NoError(t, err)
	t.Log("Out 2...", coins)
}
