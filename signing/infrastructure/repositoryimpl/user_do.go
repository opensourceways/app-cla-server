package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

const fieldAccount = "account"

// userDO
type userDO struct {
	Email          string `bson:"email"     json:"email"     required:"true"`
	LinkId         string `bson:"link_id"   json:"link_id"   required:"true"`
	Account        string `bson:"account"   json:"account"   required:"true"`
	Password       string `bson:"password"  json:"password"  required:"true"`
	CorpSigningId  string `bson:"cs_id"     json:"cs_id"     required:"true"`
	PasswordChaged bool   `bson:"changed"   json:"changed"`
	Version        int    `bson:"version"   json:"-"`
}

func (do *userDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func toUserDO(u *domain.User) userDO {
	return userDO{}
}
