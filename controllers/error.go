package controllers

import "github.com/opensourceways/app-cla-server/dbmodels"

const (
	ErrorStart = iota
	ErrInvalidParameter
	ErrHasSigned
	ErrMissingToken
	ErrUnknownToken
	ErrInvalidToken
)

func convertDBError(err error) (int, int) {
	if dbmodels.IsInvalidParameter(err) {
		return 400, ErrInvalidParameter

	} else if dbmodels.IsHasSigned(err) {
		return 400, ErrHasSigned
	}

	return 500, 0
}
