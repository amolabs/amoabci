package tx

import "encoding/json"

func classifyTxV6(base TxBase) Tx {
	var t Tx
	// TODO: use err return from parseSomethingParam()
	switch base.Type {
	case "transfer":
		param, _ := parseTransferParamV5(base.Payload)
		t = &TxTransferV5{
			TxBase: base,
			Param:  param,
		}
	case "stake":
		param, _ := parseStakeParam(base.Payload)
		t = &TxStake{
			TxBase: base,
			Param:  param,
		}
	case "withdraw":
		param, _ := parseWithdrawParam(base.Payload)
		t = &TxWithdraw{
			TxBase: base,
			Param:  param,
		}
	case "delegate":
		param, _ := parseDelegateParam(base.Payload)
		t = &TxDelegate{
			TxBase: base,
			Param:  param,
		}
	case "retract":
		param, _ := parseRetractParam(base.Payload)
		t = &TxRetract{
			TxBase: base,
			Param:  param,
		}
	case "setup":
		param, _ := parseSetupParam(base.Payload)
		t = &TxSetup{
			TxBase: base,
			Param:  param,
		}
	case "close":
		param, _ := parseCloseParam(base.Payload)
		t = &TxClose{
			TxBase: base,
			Param:  param,
		}
	case "register":
		param, _ := parseRegisterParam(base.Payload)
		t = &TxRegister{
			TxBase: base,
			Param:  param,
		}
	case "discard":
		param, _ := parseDiscardParam(base.Payload)
		t = &TxDiscard{
			TxBase: base,
			Param:  param,
		}
	case "request":
		param, _ := parseRequestParam(base.Payload)
		t = &TxRequest{
			TxBase: base,
			Param:  param,
		}
	case "cancel":
		param, _ := parseCancelParam(base.Payload)
		t = &TxCancel{
			TxBase: base,
			Param:  param,
		}
	case "grant":
		param, _ := parseGrantParam(base.Payload)
		t = &TxGrant{
			TxBase: base,
			Param:  param,
		}
	case "revoke":
		param, _ := parseRevokeParam(base.Payload)
		t = &TxRevoke{
			TxBase: base,
			Param:  param,
		}
	case "claim":
		param, _ := parseClaimParam(base.Payload)
		t = &TxClaim{
			TxBase: base,
			Param:  param,
		}
	case "dismiss":
		param, _ := parseDismissParam(base.Payload)
		t = &TxDismiss{
			TxBase: base,
			Param:  param,
		}
	case "did.claim":
		param, _ := parseDIDClaimParam(base.Payload)
		t = &TxDIDClaim{
			TxBase: base,
			Param:  param,
		}
	case "did.dismiss":
		param, _ := parseDIDDismissParam(base.Payload)
		t = &TxDIDDismiss{
			TxBase: base,
			Param:  param,
		}
	case "issue":
		param, _ := parseIssueParam(base.Payload)
		t = &TxIssue{
			TxBase: base,
			Param:  param,
		}
	case "propose":
		param, _ := parseProposeParam(base.Payload)
		t = &TxPropose{
			TxBase: base,
			Param:  param,
		}
	case "vote":
		param, _ := parseVoteParam(base.Payload)
		t = &TxVote{
			TxBase: base,
			Param:  param,
		}
	case "lock":
		param, _ := parseLockParam(base.Payload)
		t = &TxLock{
			TxBase: base,
			Param:  param,
		}
	case "burn":
		param, _ := parseBurnParam(base.Payload)
		t = &TxBurn{
			TxBase: base,
			Param:  param,
		}
	default:
		t = &base
	}
	return t
}

func ParseTxV6(txBytes []byte) (Tx, error) {
	var base TxBase

	err := json.Unmarshal(txBytes, &base)
	if err != nil {
		return nil, err
	}

	return classifyTxV6(base), nil
}
