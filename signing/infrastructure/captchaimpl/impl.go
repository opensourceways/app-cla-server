package captchaimpl

import (
	"errors"

	"github.com/mojocn/base64Captcha"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

var errCaptchaInvalid = domain.NewDomainError(domain.ErrorCodeCaptchaInvalid)

// NewCaptchaImpl creates a CaptchaService implementation.
// In test mode, every generated captcha has a fixed answer (testAnswer).
func NewCaptchaImpl(d dao, cfg *Config, isTestEnv bool, testAnswer string) *captchaImpl {
	store := &captchaStore{dao: d, expiry: cfg.expiry()}

	driver := base64Captcha.NewDriverDigit(
		80,  // height
		240, // width
		6,   // digit length
		0.7, // noise level
		80,  // background circle count
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
	// if c.isTestEnv {
	// 	id = uuid.New().String()
	// 	if err = c.store.Set(id, c.testAnswer); err != nil {
	// 		return
	// 	}
	// 	imageBase64 = "test_mode"
	// 	return
	// }

	id, imageBase64, _, err = c.captcha.Generate()
	return
}

// Verify validates the captcha answer. The captcha is consumed on success.
func (c *captchaImpl) Verify(id string, answer string) error {
	if id == "" || answer == "" {
		return errCaptchaInvalid
	}

	if !c.captcha.Verify(id, answer, true) {
		return errors.New(string(errCaptchaInvalid))
	}

	return nil
}
