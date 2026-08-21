package captchaservice

// CaptchaService handles graphic captcha generation and verification
// to prevent brute-force login attacks.
type CaptchaService interface {
	// Generate creates a new captcha and returns its ID and base64-encoded image.
	Generate() (id string, imageBase64 string, err error)

	// Verify checks whether the provided answer matches the captcha for the given ID.
	// Returns an error if the captcha is invalid or expired.
	Verify(id string, answer string) error
}
