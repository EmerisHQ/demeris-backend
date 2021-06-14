package tx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	router.GET("/tx/ticket/:ticket", GetTicket)
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

	data, err := hex.DecodeString(txRequest.TxBytes)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("failed to decode tx bytes"), http.StatusBadRequest)

		d.WriteError(c, e,
			"Failed to decode tx bytes",
			"id",
			e.ID,
			"error",
			err,
		)
	}

	tx := sdktx.Tx{}

	d.Codec.MustUnmarshalBinaryBare(data, &tx)

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

	err = validateTx(&tx, &meta, d)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("invalid transaction"), http.StatusBadRequest)

		d.WriteError(c, e,
			"invalid transaction",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	txhash, err := relayTx(d, tx, meta)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("relaying tx failed"), http.StatusBadRequest)

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

	err := validateSignatures(tx, meta, d)

	if err != nil {
		return err
	}

	err = validateBody(tx, meta, d)

	if err != nil {
		return err
	}

	// TODO: Fetch sequence for ticketing system
	err = validateAuthInfo(tx, meta, d)

	if err != nil {
		return err
	}

	return nil
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

				channels := strings.Split(denomTrace.Path, "/transfer")
				if channels[0] != sourceChannel {
					return fmt.Errorf("IBC forward is disabled for multi-hop tokens. Try sending it back through the original channel.")
				}
			}
		}
	}

	return nil
}

// validateAuthInfo validates the data inside auth_info and populates the relevant metadata
func validateAuthInfo(tx *sdktx.Tx, meta *TxMeta, d *deps.Deps) error {

	if infos := tx.AuthInfo.SignerInfos; len(infos) == 1 {
		// Fetch signer sequence
		meta.SignerSequence = string(tx.AuthInfo.SignerInfos[0].Sequence)
	} else {
		return fmt.Errorf("Invalid number of signatures. Expected 1, got %d", len(infos))
	}

	return nil
}

// validateSignatures ensures the signature exists
func validateSignatures(tx *sdktx.Tx, meta *TxMeta, d *deps.Deps) error {
	if len(tx.Signatures) != 1 {
		return fmt.Errorf("Invalid number of signatures")
	}

	return nil
}

// RelayTx relays the tx to the specifc endpoint
// RelayTx will also perform the ticketing mechanism
// Always expect broadcast mode to be `async`
func relayTx(d *deps.Deps, tx sdktx.Tx, meta TxMeta) (string, error) {

	b := d.Codec.MustMarshalBinaryBare(&tx)

	grpcConn, err := grpc.Dial(
		meta.Chain.NodeInfo.Endpoint, // Or your gRPC server address.
		grpc.WithInsecure(),          // The SDK doesn't support any transport security mechanism.
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
			TxBytes: b, // Proto-binary of the signed transaction, see previous step.
		},
	)
	if err != nil {
		return grpcRes.TxResponse.TxHash, err
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
// @ID chain
// @Description Gets transaction status by ticket id.
// @Param ticketId path string true "ticket id"
// @Produce json
// @Success 200 {object} TxStatus
// @Failure 500,403 {object} deps.Error
// @Router /tx/ticket/{ticketId} [get]
func GetTicket(c *gin.Context) {
	var res TxStatus

	d := deps.GetDeps(c)

	ticketId := c.Param("ticketId")

	ticket, err := d.Store.Get(ticketId)

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

	json.Unmarshal([]byte(ticket), &res)

	c.JSON(http.StatusOK, res)
}
