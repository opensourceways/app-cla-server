package randomstrimpl

func NewRandomStrImpl() *randomStrImpl {
	return &randomStrImpl{}
}

type randomStrImpl struct{}

func (impl *randomStrImpl) New() (string, error) { return "", nil }
