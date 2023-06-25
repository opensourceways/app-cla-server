package userpassword

type UserPassword interface {
	New() (string, error)
	IsValid(string) bool
}
