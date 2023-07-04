package encryption

type Encryption interface {
	Encrypt([]byte) ([]byte, error)
	IsSame(plainText []byte, encrypted []byte) bool
}
