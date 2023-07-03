package encryption

type Encryption interface {
	Encrypt(string) (string, error)
}
