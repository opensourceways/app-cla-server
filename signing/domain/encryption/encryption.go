package encryption

type Encryption interface {
	Ecrypt(string) (string, error)
}
