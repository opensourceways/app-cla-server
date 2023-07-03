package repository

type AccessToken interface {
	Add([]byte) (string, error)
	Find(string) ([]byte, error)
	Delete(string) error
}
