package encryptionimpl

func NewEncryptionImpl() *encryptionImpl {
	return &encryptionImpl{}
}

type encryptionImpl struct{}

func (impl *encryptionImpl) Ecrypt(string) (string, error) { return "", nil }
