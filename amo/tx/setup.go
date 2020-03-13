package tx

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/amolabs/amoabci/amo/code"
	"github.com/amolabs/amoabci/amo/store"
	"github.com/amolabs/amoabci/amo/types"
)

type SetupParam struct {
	Storage         uint32         `json:"storage"`
	Url             string         `json:"url"`
	RegistrationFee types.Currency `json:"registration_fee"`
	HostingFee      types.Currency `json:"hosting_fee"`
}

func parseSetupParam(raw []byte) (SetupParam, error) {
	var param SetupParam
	err := json.Unmarshal(raw, &param)
	if err != nil {
		return param, err
	}
	return param, nil
}

type TxSetup struct {
	TxBase
	Param SetupParam `json:"-"`
}

var _ Tx = &TxSetup{}

func (t *TxSetup) Check() (uint32, string) {
	// TOOD: check url format
	_, err := parseSetupParam(t.getPayload())
	if err != nil {
		return code.TxCodeBadParam, err.Error()
	}

	return code.TxCodeOK, "ok"
}

func (t *TxSetup) Execute(s *store.Store) (uint32, string, []abci.Event) {
	param := t.Param
	sender := t.GetSender()

	sto := s.GetStorage(param.Storage, false)
	if sto == nil {
		sto = &types.Storage{
			Owner:           sender,
			Url:             param.Url,
			RegistrationFee: param.RegistrationFee,
			HostingFee:      param.HostingFee,
			Active:          true,
		}
	} else {
		if bytes.Equal(sender, sto.Owner) == false {
			return code.TxCodePermissionDenied, "permission denied", nil
		}
		// update fields
		sto.Url = param.Url
		sto.RegistrationFee = param.RegistrationFee
		sto.HostingFee = param.HostingFee
		sto.Active = true
	}
	// store
	err := s.SetStorage(param.Storage, sto)
	if err != nil {
		return code.TxCodeUnknown, err.Error(), nil
	}
	return code.TxCodeOK, "ok", nil
}
