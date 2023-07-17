package userpassword

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type UserPassword interface {
	New() (dp.Password, error)
	IsValid(dp.Password) bool
}
