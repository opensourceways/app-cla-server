package models

import (
	"errors"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type PasswordRetrieval struct {
	Password []byte `json:"password"`
}

type PasswordRetrievalKey struct {
	Email  string `json:"email"     required:"true"`
	LinkId string `json:"link_id"`
}

func (p *PasswordRetrievalKey) Validate() IModelError {
	if p.Email == "" {
		return newModelError(ErrMissgingEmail, errors.New("missing email"))
	}

	if p.LinkId == "" {
		p.LinkId = domain.CommunityManagerLinkId()
	}

	return nil
}

func (p *PasswordRetrievalKey) IsCommunityManager() bool {
	return p.LinkId == domain.CommunityManagerLinkId()
}
