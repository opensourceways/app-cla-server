package models

import (
	"fmt"
	"regexp"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func checkEmailFormat(email string) (string, error) {
	rg := regexp.MustCompile("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")
	if !rg.MatchString(email) {
		return util.ErrInvalidEmail, fmt.Errorf("invalid email:%s", email)
	}

	return "", nil
}

func parseErrorOfDBApi(err error) (string, error) {
	if err == nil {
		return "", err
	}

	if e, ok := dbmodels.IsDBError(err); ok {
		return e.ErrCode, e.Err
	}

	return util.ErrSystemError, err
}

func isNoDBRecord(err error) bool {
	e, ok := dbmodels.IsDBError(err)
	return ok && e.ErrCode == util.ErrNoDBRecord
}
