package client

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateClient(t *testing.T) {
	cli, err := New(t, context.Background())
	require.NoError(t, err)

	// accountName := "test5"

	// account, _, err := cli.CreateAccount(accountName)
	// if err != nil {
	// 	log.Println("create account error....", err)
	// }

	// _ = account

	//a, err := cli.ImportMnemonic("c1", "foot milk eight ugly nation atom deer tuition door quarter tackle bicycle three fall purpose behave school shy tonight decrease local concert snap false")
	a, err := cli.ImportMnemonic("test_key", "unable grass pipe pear glue forward gold truck victory pause fragile scan mask morning develop floor treat essence vendor solid surprise theme process sign")
	if err != nil {
		log.Println("mnemonic import error.....", err)
	}

	log.Println("aaaaaaaaaaaaaaaaa.......", a)

	// get account from the keyring by account name and return a bech32 address
	address, err := cli.Address("test_key")
	if err != nil {
		log.Println("get address error", err)
	}

	list, err := cli.GetkeysList()
	if err != nil {
		fmt.Println("Error while getting keys list")
	}

	fmt.Println("test_key address........", address)

	fmt.Println("list......", list[0].GetAddress().String(), list[0].GetName())
	// fmt.Println("list......", list[1].GetAddress().String(), list[1].GetName())

	adr := list[0].GetAddress().String()

	bal, err := cli.GetBankBalances(adr, "uatom")
	if err != nil {
		fmt.Println("Error while getting bank balance....", err)
	}

	fmt.Println("balllllll..........", bal, err)

	cli.TestGetBalanceOfAnyAccount(t)

	log.Fatal("Address.....", address)
}
