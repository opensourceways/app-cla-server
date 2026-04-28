package captchaimpl

import "time"

// Config holds captcha storage and generation configuration.
type Config struct {
	// Expiry is how long a captcha remains valid, in seconds.
	Expiry int64 `json:"expiry"`
	// Height of the captcha image in pixels.
	Height int `json:"height"`
	// Width of the captcha image in pixels.
	Width int `json:"width"`
	// Length is the number of digits in the captcha.
	Length int `json:"length"`
	// NoiseLevel controls the amount of noise in the image (0-1).
	NoiseLevel float64 `json:"noise_level"`
	// BackgroundCircles is the number of background circles.
	BackgroundCircles int `json:"background_circles"`
}

func (cfg *Config) SetDefault() {
	if cfg.Expiry <= 0 {
		cfg.Expiry = 300 // 5 minutes default
	}
	if cfg.Height <= 0 {
		cfg.Height = 80
	}
	if cfg.Width <= 0 {
		cfg.Width = 240
	}
	if cfg.Length <= 0 {
		cfg.Length = 6
	}
	if cfg.NoiseLevel <= 0 {
		cfg.NoiseLevel = 0.7
	}
	if cfg.BackgroundCircles <= 0 {
		cfg.BackgroundCircles = 80
	}
}

func (cfg *Config) expiry() time.Duration {
	return time.Duration(cfg.Expiry) * time.Second
}
