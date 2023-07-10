package models

import (
	"fmt"
	"regexp"
)

func checkEmailFormat(email string) IModelError {
	rg := regexp.MustCompile("^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")
	if !rg.MatchString(email) {
		return newModelError(ErrNotAnEmail, fmt.Errorf("invalid email:%s", email))
	}
	return nil
}
