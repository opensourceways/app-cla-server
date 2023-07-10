package limiterimpl

import "time"

func NewLimiterImpl(d dao) *limiterImpl {
	return &limiterImpl{
		dao: d,
	}
}

type limiterImpl struct {
	dao dao
}

func (impl *limiterImpl) Add(k string, expiry time.Duration) error {
	return impl.dao.SetKey(k, expiry)
}

func (impl *limiterImpl) IsAllowed(k string) (bool, error) {
	return impl.dao.HasKey(k)
}
