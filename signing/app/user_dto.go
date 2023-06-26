package app

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

// CmdToChangePassword
type CmdToChangePassword struct {
	Id     string
	OldOne dp.Password
	NewOne dp.Password
}
