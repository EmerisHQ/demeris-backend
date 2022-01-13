package client

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateClient(t *testing.T) {
	cli, err := New("cosmos-hub", t, context.Background())
	require.NoError(t, err)

	a, err := cli.ImportMnemonic("c1", "foot milk eight ugly nation atom deer tuition door quarter tackle bicycle three fall purpose behave school shy tonight decrease local concert snap false")
	if err != nil {
		log.Println("mnemonic import error.....", err)
	}

	// get account from the keyring by account name and return a bech32 address
	address, err := cli.Address(a.Info.GetName())
	if err != nil {
		log.Println("get address error", err)
	}

	log.Printf("address : %v", address)

	list, err := cli.GetkeysList()
	if err != nil {
		fmt.Println("Error while getting keys list")
	}

	log.Printf("list length : %v", len(list))

	adr := list[0].GetAddress().String()

	bal, err := cli.GetBankBalances(adr, "uatom")
	if err != nil {
		fmt.Println("Error while getting bank balance....", err)
	}

	log.Printf("balance : %v", bal)

	cli.TestGetBalanceOfAnyAccount(t)
}
