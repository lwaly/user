package models

import "encoding/json"

// Predefined model error codes.
const (
	ErrDatabase = -1
	ErrSystem   = -2
	ErrDupRows  = -3
	ErrNotFound = -4
	ErrInput    = -5
)

// CodeInfo definiton.
type CodeInfo struct {
	Code int             `json:"Code"`
	Info string          `json:"Info"`
	Body json.RawMessage `json:"body"`
}

// NewErrorInfo return a CodeInfo represents error.
func NewErrorInfo(code int, info string) *CodeInfo {
	var msg []byte
	return &CodeInfo{code, info, msg}
}

// NewNormalInfo return a CodeInfo represents OK.
func NewNormalInfo(msg []byte) *CodeInfo {
	return &CodeInfo{0, "ok", msg}
}

type UserInfo struct {
	Id int `json:"Id"`
}
