package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	fieldToken = "token"
)

func toEmailCredentialDO(e *domain.EmailCredential) emailCredentialDO {
	return emailCredentialDO{
		Platform: e.Platform,
		Email:    e.Addr.EmailAddr(),
		Token:    e.Token,
	}
}

type emailCredentialDO struct {
	Platform string `bson:"platform"  json:"platform" required:"true"`
	Email    string `bson:"email"     json:"email"    required:"true"`
	Token    []byte `bson:"token"     json:"-"`
}

func (do *emailCredentialDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *emailCredentialDO) toEmailCredential() (r domain.EmailCredential, err error) {
	if r.Addr, err = dp.NewEmailAddr(do.Email); err != nil {
		return
	}

	r.Platform = do.Platform
	r.Token = do.Token

	return
}
