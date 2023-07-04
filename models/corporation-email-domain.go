package models

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

const minNumOfEmailDomainParts = 2

type CorpEmailDomainCreateOption struct {
	SubEmail         string `json:"sub_email"`
	VerificationCode string `json:"verification_code"`
}

func (cse *CorpEmailDomainCreateOption) Check(csId string) IModelError {
	if err := checkEmailFormat(cse.SubEmail); err != nil {
		return err
	}

	return validateCodeForAddingEmailDomain(
		csId, cse.SubEmail, cse.VerificationCode,
	)
}

func (cse CorpEmailDomainCreateOption) Create(linkID, adminEmail string) IModelError {
	err := dbmodels.GetDB().AddCorpEmailDomain(linkID, adminEmail, util.EmailSuffix(cse.SubEmail))
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrUnsigned, err)
	}
	return parseDBError(err)
}

func ListCorpEmailDomain(linkID, email string) ([]string, IModelError) {
	v, err := dbmodels.GetDB().GetCorpEmailDomains(linkID, email)
	if err == nil {
		if v == nil {
			v = []string{}
		}
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}

func AddCorpEmailDomain(csId string, opt *CorpEmailDomainCreateOption) IModelError {
	return corpEmailDomainAdapterInstance.Add(csId, opt)
}

func ListCorpEmailDomains(csId string) ([]string, IModelError) {
	return corpEmailDomainAdapterInstance.List(csId)
}

func isMatchedEmailDomain(email1, email2 string) bool {
	e1 := strings.Split(util.EmailSuffix(email1), ".")
	e2 := strings.Split(util.EmailSuffix(email2), ".")
	n1 := len(e1) - 1
	j := len(e2) - 1
	for i := n1; i >= 0; i-- {
		if j < 0 {
			break
		}
		if e1[i] != e2[j] {
			return n1-i >= minNumOfEmailDomainParts
		}

		j--
	}
	return true
}

func PurposeOfAddingEmailDomain(csId string) string {
	return fmt.Sprintf("adding email domain:%s", csId)
}
