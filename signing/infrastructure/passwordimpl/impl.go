package passwordimpl

func NewPasswordImpl() *passwordImpl {
	return &passwordImpl{}
}

type passwordImpl struct{}

func (impl *passwordImpl) New() (string, error) { return "", nil }

func (impl *passwordImpl) IsValid(string) bool { return true }
