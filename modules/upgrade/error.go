package upgrade

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = 1

	CodeInvalidMsgType			sdk.CodeType = 200
	CodeUnknownRequest  		sdk.CodeType = sdk.CodeUnknownRequest
)

func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case CodeInvalidMsgType:
		return "Invalid msg type"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

func NewError(codespace sdk.CodespaceType, code sdk.CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}

func msgOrDefaultMsg(msg string, code sdk.CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}