package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	fieldAccount    = "account"
	fieldChanged    = "changed"
	fieldPrivacy    = "privacy"
	fieldPassword   = "password"
	fieldFailedNum  = "failed_num"
	fieldLoginTime  = "login_time"
	fieldFrozenTime = "frozen_time"
)

type privacyConsentDO struct {
	Time    string `bson:"time"     json:"time"     required:"true"`
	Version string `bson:"version"  json:"version"  required:"true"`
}

func (do *privacyConsentDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

// userDO
type userDO struct {
	Id              primitive.ObjectID `bson:"_id"           json:"-"`
	Email           string             `bson:"email"         json:"email"     required:"true"`
	LinkId          string             `bson:"link_id"       json:"link_id"   required:"true"`
	Account         string             `bson:"account"       json:"account"   required:"true"`
	Password        []byte             `bson:"password"      json:"-"`
	CorpSigningId   string             `bson:"cs_id"         json:"cs_id"     required:"true"`
	PrivacyConsent  privacyConsentDO   `bson:"privacy"       json:"privacy"`
	PasswordChanged bool               `bson:"changed"       json:"changed"`
	Version         int                `bson:"version"       json:"-"`
}

func (do *userDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *userDO) toUser() domain.User {
	return domain.User{
		LinkId:        do.LinkId,
		CorpSigningId: do.CorpSigningId,
		UserBasicInfo: domain.UserBasicInfo{
			Id:              do.Id.Hex(),
			Account:         dp.CreateAccount(do.Account),
			Password:        do.Password,
			EmailAddr:       dp.CreateEmailAddr(do.Email),
			PasswordChanged: do.PasswordChanged,
			PrivacyConsent: domain.PrivacyConsent{
				Time:    do.PrivacyConsent.Time,
				Version: do.PrivacyConsent.Version,
			},
			Version: do.Version,
		},
	}
}

func toUserDO(u *domain.User) userDO {
	return userDO{
		Email:           u.EmailAddr.EmailAddr(),
		LinkId:          u.LinkId,
		Account:         u.Account.Account(),
		CorpSigningId:   u.CorpSigningId,
		PasswordChanged: u.PasswordChanged,
	}
}
