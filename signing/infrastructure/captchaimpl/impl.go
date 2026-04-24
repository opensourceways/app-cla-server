package captchaimpl

import (
	"github.com/mojocn/base64Captcha"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

var errCaptchaInvalid = domain.NewDomainError(domain.ErrorCodeCaptchaInvalid)

// NewCaptchaImpl creates a CaptchaService implementation.
// In test mode, every generated captcha has a fixed answer (testAnswer).
func NewCaptchaImpl(d dao, cfg *Config, isTestEnv bool, testAnswer string) *captchaImpl {
	store := &captchaStore{dao: d, expiry: cfg.expiry()}

	driver := base64Captcha.NewDriverDigit(
		cfg.Height,
		cfg.Width,
		cfg.Length,
		cfg.NoiseLevel,
		cfg.BackgroundCircles,
	)

	return &captchaImpl{
		captcha:    base64Captcha.NewCaptcha(driver, store),
		store:      store,
		isTestEnv:  isTestEnv,
		testAnswer: testAnswer,
	}
}

type captchaImpl struct {
	captcha    *base64Captcha.Captcha
	store      *captchaStore
	isTestEnv  bool
	testAnswer string
}

// Generate creates a new captcha and returns (id, base64Image, error).
func (c *captchaImpl) Generate() (id string, imageBase64 string, err error) {
	id, imageBase64, _, err = c.captcha.Generate()
	return
}

// Verify validates the captcha answer. The captcha is consumed on success.
func (c *captchaImpl) Verify(id string, answer string) error {
	if id == "" || answer == "" {
		return errCaptchaInvalid
	}

	// TODO: uncomment to enable test mode
	// if c.isTestEnv && answer == c.testAnswer {
	// 	c.store.Get(id, true)
	// 	return nil
	// }

	if !c.captcha.Verify(id, answer, true) {
		return errCaptchaInvalid
	}

	return nil
}
