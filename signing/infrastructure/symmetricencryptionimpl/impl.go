package symmetricencryptionimpl

func NewSymmetricEncryptionImpl() *symmetricEncryptionImpl {
	return &symmetricEncryptionImpl{}
}

type symmetricEncryptionImpl struct{}

func (impl *symmetricEncryptionImpl) Encrypt(plaintext []byte) ([]byte, error) {
	return nil, nil
}

func (impl *symmetricEncryptionImpl) Decrypt(ciphertext []byte) ([]byte, error) {
	return nil, nil
}
