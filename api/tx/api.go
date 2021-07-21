package tx

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	types2 "github.com/tendermint/tendermint/abci/types"

	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	"google.golang.org/grpc"

	"github.com/allinbits/demeris-backend/api/router/deps"
	// typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gin-gonic/gin"
	// "google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/types/known/anypb"
)

func Register(router *gin.Engine) {
	router.POST("/tx/:chain", Tx)
	router.GET("/tx/ticket/:chain/:ticket", GetTicket)
}

// Tx relays a transaction to an internal node for the specified chain.
// @Summary Relays a transaction to the relevant chain.
// @Tags Tx
// @ID tx
// @Description Relays a transaction to the relevant chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} TxResponse
// @Failure 500,403 {object} deps.Error
// @Router /tx/{chainName} [post]
func Tx(c *gin.Context) {
	// var tx typestx.Tx
	var txRequest TxRequest
	var meta TxMeta

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	err := c.BindJSON(&txRequest)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("failed to parse JSON"), http.StatusBadRequest)

		d.WriteError(c, e,
			"Failed to parse JSON",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	tx := sdktx.Tx{}

	mustCheckTx := true
	if err := d.Codec.UnmarshalBinaryBare(txRequest.TxBytes, &tx); err != nil {
		mustCheckTx = false
		d.Logger.Warnw("cannot decode transaction with the codec we have, bypassing transaction checking", "error", err)
	}

	meta.Chain, err = d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("chain %s does not exist", chainName), http.StatusBadRequest)

		d.WriteError(c, e,
			"Invalid chain",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	var validationErr error

	if mustCheckTx {
		validationErr = validateTx(&tx, &meta, d)
	}

	if validationErr != nil {
		e := deps.NewError("tx", fmt.Errorf("invalid transaction"), http.StatusBadRequest)

		d.WriteError(c, e,
			"invalid transaction",
			"id",
			e.ID,
			"error",
			validationErr,
		)

		return
	}

	txhash, err := relayTx(d, txRequest.TxBytes, meta)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("relaying tx failed, %w", err), http.StatusBadRequest)

		d.WriteError(c, e,
			"relaying tx failed",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, TxResponse{
		Ticket: txhash,
	})
}

// validateTx populates metadata and
func validateTx(tx *sdktx.Tx, meta *TxMeta, d *deps.Deps) error {
	return validateBody(tx, meta, d)
}

// validateBody validates the data inside the body and populates the relevant metadata
func validateBody(tx *sdktx.Tx, meta *TxMeta, d *deps.Deps) error {
	for _, m := range tx.GetMsgs() {
		if m.Type() == "transfer" {

			msg, ok := m.(*types.MsgTransfer)

			if !ok {
				return fmt.Errorf("expected MsgTransfer, got %T", msg)
			}

			sourcePort := msg.SourcePort
			sourceChannel := msg.SourceChannel

			tokenDenom := msg.Token.Denom

			fmt.Println(sourcePort, sourceChannel, tokenDenom)

			if sourcePort != "transfer" {
				return fmt.Errorf("Invalid IBC Port %s", sourcePort)
			}

			if "ibc/" == tokenDenom[:4] {
				tokenHash := tokenDenom[4:]
				denomTrace, err := d.Database.DenomTrace(meta.Chain.ChainName, tokenHash)
				if err != nil {
					return fmt.Errorf("Invalid denom trace")
				}

				// TODO: move this to the chains/api.go.path() function
				channels := strings.Split(denomTrace.Path, "/")
				if channels[1] != sourceChannel {
					return fmt.Errorf("IBC forward is disabled for multi-hop tokens. Try sending it back through the original channel.")
				}
			}
		}
	}

	return nil
}

// RelayTx relays the tx to the specifc endpoint
// RelayTx will also perform the ticketing mechanism
// Always expect broadcast mode to be `async`
func relayTx(d *deps.Deps, txBytes []byte, meta TxMeta) (string, error) {

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", meta.Chain.ChainName, 9090), // Or your gRPC server address.
		grpc.WithInsecure(), // The SDK doesn't support any transport security mechanism.
	)

	if err != nil {
		return "", fmt.Errorf("cannot create grpc dialer, %w", err)
	}

	defer grpcConn.Close()

	txClient := sdktx.NewServiceClient(grpcConn)
	// We then call the BroadcastTx method on this client.
	grpcRes, err := txClient.BroadcastTx(
		context.Background(),
		&sdktx.BroadcastTxRequest{
			Mode:    sdktx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		},
	)

	if err != nil {
		return "", err
	}

	msgs := grpcRes.TxResponse.GetTx().GetMsgs()
	signers := msgs[0].GetSigners()
	if len(signers) != 1{
		// TODO: add error
		return
	}

	// TODO pass signer to create key

	if grpcRes.TxResponse.Code != types2.CodeTypeOK {
		return "", fmt.Errorf("transaction relaying error: code %d, %s", grpcRes.TxResponse.Code, grpcRes.TxResponse.RawLog)
	}

	err = d.Store.CreateTicket(meta.Chain.ChainName, grpcRes.TxResponse.TxHash)

	if err != nil {
		return grpcRes.TxResponse.TxHash, err
	}

	return grpcRes.TxResponse.TxHash, nil
}

// GetTicket returns the transaction status n.
// @Summary Gets ticket by id.
// @Tags Chain
// @ID txTicket
// @Description Gets transaction status by ticket id.
// @Param ticketId path string true "ticket id"
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} store.Ticket
// @Failure 500,403 {object} deps.Error
// @Router /tx/ticket/{chainName}/{ticketId} [get]
func GetTicket(c *gin.Context) {

	d := deps.GetDeps(c)

	chainName := c.Param("chain")
	ticketId := c.Param("ticket")

	ticket, err := d.Store.Get(fmt.Sprintf("%s-%s", chainName, ticketId))

	if err != nil {
		e := deps.NewError(
			"tx",
			fmt.Errorf("cannot retrieve ticket with id %v", ticketId),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve ticket",
			"id",
			e.ID,
			"name",
			ticketId,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, ticket)
}
