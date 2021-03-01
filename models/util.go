package models

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
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

func encryptPassword(pwd string) (string, IModelError) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", newModelError(ErrSystemError, err)
	}

	return string(hash), nil
}

func isSamePasswords(hashedPwd, plainPwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd)) == nil
}
