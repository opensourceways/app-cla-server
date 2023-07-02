package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type EmailCredential struct {
	Addr     dp.EmailAddr
	Token    []byte
	Platform string
}
