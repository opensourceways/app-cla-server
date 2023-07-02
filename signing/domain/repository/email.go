package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type EmailCredential interface {
	Add(*domain.EmailCredential) error
	Find(dp.EmailAddr) (domain.EmailCredential, error)
}
