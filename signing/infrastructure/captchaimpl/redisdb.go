package captchaimpl

import (
	"strings"
	"time"
)

// dao is the Redis DAO interface required by captchaStore.
type dao interface {
	SetWithExpiry(key string, val interface{}, expiry time.Duration) error
	Get(key string, val interface{}) error
	Expire(key string, expire time.Duration) error
	IsDocNotExists(err error) bool
}

const captchaKeyPrefix = "captcha:"

func captchaKey(id string) string {
	return captchaKeyPrefix + id
}

// captchaStore implements base64Captcha.Store backed by Redis.
type captchaStore struct {
	dao    dao
	expiry time.Duration
}

// Set stores the captcha answer in Redis with TTL.
func (s *captchaStore) Set(id string, value string) error {
	return s.dao.SetWithExpiry(captchaKey(id), value, s.expiry)
}

// Get retrieves the captcha answer from Redis and optionally deletes it.
func (s *captchaStore) Get(id string, clear bool) string {
	var val string
	if err := s.dao.Get(captchaKey(id), &val); err != nil {
		return ""
	}

	if clear {
		// expire immediately
		_ = s.dao.Expire(captchaKey(id), 0)
	}

	return val
}

// Verify checks the captcha answer (case-insensitive).
func (s *captchaStore) Verify(id, answer string, clear bool) bool {
	stored := s.Get(id, clear)
	return stored != "" && strings.EqualFold(stored, answer)
}
