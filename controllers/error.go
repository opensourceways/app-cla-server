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
	e, ok := dbmodels.IsDBError(err)
	if !ok {
		return 500, 0
	}

	if dbmodels.IsInvalidParameter(e) {
		return 400, ErrInvalidParameter

	} else if dbmodels.IsHasSigned(e) {
		return 400, ErrHasSigned
	}

	return 500, 0
}
