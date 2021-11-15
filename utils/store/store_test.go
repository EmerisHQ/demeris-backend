package store

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	testChain      = "cosmos-hub"
	testTxHash     = "918DC23785CABA3EE4E4A59321E679F8B7A2E27C9DFB165B3B6D22EF23017264"
	testOwner      = "cosmos1l2lepugxx5heetsl2cs74e2sy0uqxv390as7zw"
	testDestChain  = "cosmos-hub-2"
	testSrcChannel = "channel-1"
	testPktSeq     = "1"
	testErr        = "dummy error"
)

var (
	store *Store
	mr    *miniredis.Miniredis
)

func TestMain(m *testing.M) {
	mr, store = SetupTestDB()
	code := m.Run()
	defer mr.Close()
	os.Exit(code)
}

func getShadowKey(key string) string {
	return shadow + key
}

func TestCreateTicket(t *testing.T) {
	defer ResetTestDB(mr, store)
	// create ticket
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	key := GetKey(testChain, testTxHash)
	require.True(t, store.Exists(key))
	require.True(t, store.Exists(getShadowKey(key)))
	// get all tickets of testOwner
	tickets, err := store.GetUserTickets(testOwner)
	require.NoError(t, err)
	require.Len(t, tickets[testChain], 1)
	require.Equal(t, testTxHash, tickets[testChain][0])
	// get created ticket details
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, pending, ticket.Status)
}

func TestSetComplete(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetComplete with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetComplete(key, 123))
	// create ticket and set complete
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetComplete(key, 123))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, complete, ticket.Status)
	// check shadow key deleted
	require.False(t, store.Exists(getShadowKey(key)))
	// get all tickets of testOwner
	tickets, err := store.GetUserTickets(testOwner)
	require.NoError(t, err)
	// expecting no tickets for owner
	require.Len(t, tickets[testChain], 0)
}

func TestSetInTransit(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetInTransit with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetInTransit(key, testDestChain, testSrcChannel,
		testPktSeq, testTxHash, testChain, 123))
	// create ticket and set ticket status as transit
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetInTransit(key, testDestChain, testSrcChannel,
		testPktSeq, testTxHash, testChain, 123))
	require.True(t, store.Exists(key))
	require.True(t, store.Exists(getShadowKey(key)))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, transit, ticket.Status)
	newKey := GetIBCKey(testDestChain, testSrcChannel, testPktSeq)
	require.True(t, store.Exists(newKey))
	// get created ticket details of new key
	newKeyTicket, err := store.Get(newKey)
	require.NoError(t, err)
	require.Len(t, newKeyTicket.TxHashes, 1)
}

func TestSetIbcReceived(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetIbcReceived with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetIbcReceived(key, testTxHash, testChain, 123))
	// create ticket and set ticket status as transit
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetInTransit(key, testDestChain, testSrcChannel,
		testPktSeq, testTxHash, testChain, 123))
	require.True(t, store.Exists(key))
	require.True(t, store.Exists(getShadowKey(key)))
	newKey := GetIBCKey(testDestChain, testSrcChannel, testPktSeq)
	require.True(t, store.Exists(newKey))
	// update new key ticket status to ibcReceiveSuccess
	require.NoError(t, store.SetIbcReceived(newKey, testTxHash, testChain, 144))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, ibcReceiveSuccess, ticket.Status)
	require.Len(t, ticket.TxHashes, 2)
	// check shadow key deleted
	require.False(t, store.Exists(getShadowKey(key)))
	// get all tickets of testOwner
	tickets, err := store.GetUserTickets(testOwner)
	require.NoError(t, err)
	// expecting no tickets for owner
	require.Len(t, tickets[testChain], 0)
}

func TestSetIbcFailed(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetIbcFailed with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetIbcFailed(key, testTxHash, testChain, 123))
	// create ticket and set ticket status as transit
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetInTransit(key, testDestChain, testSrcChannel,
		testPktSeq, testTxHash, testChain, 123))
	require.True(t, store.Exists(key))
	require.True(t, store.Exists(getShadowKey(key)))
	newKey := GetIBCKey(testDestChain, testSrcChannel, testPktSeq)
	require.True(t, store.Exists(newKey))
	// update new key ticket status to ibcReceiveFailed
	require.NoError(t, store.SetIbcFailed(newKey, testTxHash, testChain, 144))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, ibcReceiveFailed, ticket.Status)
	require.Len(t, ticket.TxHashes, 2)
	// check shadow key still exists
	require.True(t, store.Exists(getShadowKey(key)))
}

func TestSetIbcTimeoutUnlock(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetIbcTimeoutUnlock with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetIbcTimeoutUnlock(key, testTxHash, testChain, 123))
	// create ticket and set ticket status as transit
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetInTransit(key, testDestChain, testSrcChannel,
		testPktSeq, testTxHash, testChain, 123))
	require.True(t, store.Exists(key))
	require.True(t, store.Exists(getShadowKey(key)))
	newKey := GetIBCKey(testDestChain, testSrcChannel, testPktSeq)
	require.True(t, store.Exists(newKey))
	// update new key ticket status to tokensUnlockedTimeout
	require.NoError(t, store.SetIbcTimeoutUnlock(newKey, testTxHash, testChain, 144))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, tokensUnlockedTimeout, ticket.Status)
	require.Len(t, ticket.TxHashes, 2)
	// check shadow key deleted
	require.False(t, store.Exists(getShadowKey(key)))
	// get all tickets of testOwner
	tickets, err := store.GetUserTickets(testOwner)
	require.NoError(t, err)
	// expecting no tickets for owner
	require.Len(t, tickets[testChain], 0)
}

func TestSetIbcAckUnlock(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetIbcAckUnlock with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetIbcAckUnlock(key, testTxHash, testChain, 123))
	// create ticket and set ticket status as transit
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetInTransit(key, testDestChain, testSrcChannel,
		testPktSeq, testTxHash, testChain, 123))
	require.True(t, store.Exists(key))
	require.True(t, store.Exists(getShadowKey(key)))
	newKey := GetIBCKey(testDestChain, testSrcChannel, testPktSeq)
	require.True(t, store.Exists(newKey))
	// update new key ticket status to tokensUnlockedAck
	require.NoError(t, store.SetIbcAckUnlock(newKey, testTxHash, testChain, 144))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, tokensUnlockedAck, ticket.Status)
	require.Len(t, ticket.TxHashes, 2)
	// check shadow key deleted
	require.False(t, store.Exists(getShadowKey(key)))
	// get all tickets of testOwner
	tickets, err := store.GetUserTickets(testOwner)
	require.NoError(t, err)
	// expecting no tickets for owner
	require.Len(t, tickets[testChain], 0)
}

func TestSetFailedWithErr(t *testing.T) {
	defer ResetTestDB(mr, store)
	// call SetFailedWithErr with not stored key, expecting error
	key := GetKey(testChain, testTxHash)
	require.Error(t, store.SetFailedWithErr(key, testErr, 123))
	// create ticket and set ticket status as failed
	require.NoError(t, store.CreateTicket(testChain, testTxHash, testOwner))
	require.NoError(t, store.SetFailedWithErr(key, testErr, 123))
	require.True(t, store.Exists(key))
	// get updated ticket details of key
	ticket, err := store.Get(key)
	require.NoError(t, err)
	require.Equal(t, failed, ticket.Status)
	require.Equal(t, testErr, ticket.Error)
	// get all tickets of testOwner
	tickets, err := store.GetUserTickets(testOwner)
	require.NoError(t, err)
	// expecting no tickets for owner
	require.Len(t, tickets[testChain], 0)
}

func TestSetPoolSwapFees(t *testing.T) {
	defer ResetTestDB(mr, store)
	var (
		testPoolID = "2"
		testAmount = "1000"
		testDenom  = "stake"
	)
	// call TestSetPoolSwapFees with wrong amount format
	require.Error(t, store.SetPoolSwapFees(testPoolID, "amount", testDenom))
	// set swap fees for pool
	require.NoError(t, store.SetPoolSwapFees(testPoolID, testAmount, testDenom))
	fees, err := store.GetSwapFees(testPoolID)
	require.NoError(t, err)
	testAmountInt, ok := sdk.NewIntFromString(testAmount)
	require.True(t, ok)
	require.Equal(t, sdk.Coins{sdk.NewCoin(testDenom, testAmountInt)}.String(), fees.String())
}

func TestBlocks(t *testing.T) {
	defer ResetTestDB(mr, store)
	blocks := NewBlocks(store)
	// call Block method with height not stored, expected error
	_, err := blocks.Block(123)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrBlockNotFound)
	// add new block with dummy data
	data := []byte("dummy block data")
	require.NoError(t, blocks.Add(data, 123))
	// test Block Method
	res, err := blocks.Block(123)
	require.NoError(t, err)
	require.Equal(t, data, res)
}
