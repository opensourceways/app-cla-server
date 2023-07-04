package encryption

type Encryption interface {
	Encrypt(string) ([]byte, error)
	IsSame(plainText string, encrypted []byte) bool
}
