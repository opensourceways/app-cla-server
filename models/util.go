package models

import (
	"fmt"
	"regexp"

	"github.com/opensourceways/app-cla-server/config"
)

func checkEmailFormat(email string) IModelError {
	rg := regexp.MustCompile("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")
	if !rg.MatchString(email) {
		return newModelError(ErrNotAnEmail, fmt.Errorf("invalid email:%s", email))
	}
	return nil
}

func checkManagerID(mid string) IModelError {
	rg := regexp.MustCompile("^[a-zA-Z0-9_.-]+_[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")
	if !rg.MatchString(mid) {
		return newModelError(ErrInvalidManagerID, fmt.Errorf("invalid manager id:%s", mid))
	}
	return nil
}

func checkPassword(s string) IModelError {
	if n := len(s); n < config.AppConfig.MinLengthOfPassword || n > config.AppConfig.MaxLengthOfPassword {
		return newModelError(ErrInvalidPassword, fmt.Errorf("the length of password is invalid"))
	}

	part := make([]bool, 4)

	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			part[0] = true
		} else if c >= 'A' && c <= 'Z' {
			part[1] = true
		} else if c >= '0' && c <= '9' {
			part[2] = true
		} else {
			part[3] = true
		}
	}

	i := 0
	for _, b := range part {
		if b {
			i++
		}
	}
	if i < 3 {
		return newModelError(
			ErrInvalidPassword,
			fmt.Errorf("the password must includes three of lowercase, uppercase, digital and special character"),
		)
	}
	return nil
}
