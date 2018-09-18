package lcd

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	"github.com/irisnet/irishub/client/context"
	"github.com/irisnet/irishub/client/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"github.com/irisnet/irishub/client/bank"
)

type sendBody struct {
	Amount sdk.Coins `json:"amount"`
	BaseTx utils.BaseTx `json:"base_tx"`
}

// SendRequestHandlerFn - http request handler to send coins to a address
// nolint: gocyclo
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		bech32addr := vars["address"]
		to, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		generateOnly, err := strconv.ParseBool(vars[utils.GenerateOnly])
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		async, err := strconv.ParseBool(vars[utils.Async])
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var m sendBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		info, err := kb.Get(m.BaseTx.LocalAccountName)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		txCtx := context.TxContext{
			Codec:         cdc,
			Gas:           m.BaseTx.Gas,
			Fee:           m.BaseTx.Fees,
			ChainID:       m.BaseTx.ChainID,
			AccountNumber: m.BaseTx.AccountNumber,
			Sequence:      m.BaseTx.Sequence,
		}

		amount, err := txCtx.ParseCoins(m.Amount.String())
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		// build message
		msg := bank.BuildMsg(sdk.AccAddress(info.GetPubKey().Address()), to, amount)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}


		if generateOnly {
			utils.WriteGenerateStdTxResponse(w, txCtx, []sdk.Msg{msg})
			return
		}

		txBytes, err := txCtx.BuildAndSign(m.BaseTx.LocalAccountName, m.BaseTx.Password, []sdk.Msg{msg})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		var res interface{}
		if async {
			res, err = cliCtx.BroadcastTxAsync(txBytes)
		} else {
			res, err = cliCtx.BroadcastTx(txBytes)
		}

		output, err := cdc.MarshalJSONIndent(res, "", "  ")
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(output)
	}
}