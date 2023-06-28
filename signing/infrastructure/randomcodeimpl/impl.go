package randomcodeimpl

func NewRandomCodeImpl() *randomCodeImpl {
	return &randomCodeImpl{}
}

type randomCodeImpl struct{}

func (impl *randomCodeImpl) New() (string, error) { return "", nil }

func (impl *randomCodeImpl) IsValid(string) bool { return true }
