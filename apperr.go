package ch9329

import "errors"

var (
	ErrInvalidKey     = errors.New("error invalid key")
	ErrInvalidModifer = errors.New("error invalid modifier")
	ErrTooManyKeys    = errors.New("error too many keys")
	ErrProtocol       = errors.New("error protocol")
)
