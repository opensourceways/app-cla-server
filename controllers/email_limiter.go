package controllers

import (
	"sync"

	"github.com/opensourceways/app-cla-server/util"
)

var emailLimiter *emailLimiterImpl

func initEmailLimiter() {
	emailLimiter = &emailLimiterImpl{
		cache: make(map[string]int64),
	}
}

type emailLimiterImpl struct {
	cache map[string]int64
	lock  sync.RWMutex
}

func (impl *emailLimiterImpl) check(linkId, email string) (pass bool) {
	k := linkId + email
	now := util.Now()

	impl.lock.RLock()
	if !impl.isAllowed(k, now) {
		impl.lock.RUnlock()

		return
	}
	impl.lock.RUnlock()

	impl.lock.Lock()
	if impl.isAllowed(k, now) {
		impl.cache[k] = now + 60

		impl.clean(now)

		pass = true
	}
	impl.lock.Unlock()

	return
}

func (impl *emailLimiterImpl) isAllowed(k string, now int64) bool {
	v, ok := impl.cache[k]

	return !ok || v <= now
}

func (impl *emailLimiterImpl) clean(now int64) {
	for k, v := range impl.cache {
		if v <= now {
			delete(impl.cache, k)
		}
	}
}
