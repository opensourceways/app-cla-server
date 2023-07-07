package models

import (
	"fmt"
	"strings"

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
