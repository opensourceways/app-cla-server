package models

import (
	"fmt"
	"regexp"

	"github.com/opensourceways/app-cla-server/util"
)

func checkEmailFormat(email string) *ModelError {
	rg := regexp.MustCompile("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")
	if !rg.MatchString(email) {
		return newModelError(ErrNotAnEmail, fmt.Errorf("invalid email:%s", email))
	}

	return nil
}

func checkManagerID(mid string) (string, error) {
	rg := regexp.MustCompile("^[a-zA-Z0-9_.-]+_[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")
	if !rg.MatchString(mid) {
		return util.ErrInvalidManagerID, fmt.Errorf("invalid manager id:%s", mid)
	}

	return "", nil
}
