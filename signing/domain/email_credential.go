package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type EmailCredential struct {
	Addr     dp.EmailAddr
	Token    []byte
	Platform string
}

func (e *EmailCredential) Clear() {
	for i := range e.Token {
		e.Token[i] = 0
	}
}
