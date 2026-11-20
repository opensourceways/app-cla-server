package controllers

import (
	"fmt"

	"github.com/opensourceways/obs-sdk/go/log"
)

// vcRateLimiter limits verification-code requests per client IP within a
// rolling time window.
type vcRateLimiter interface {
	// Allow checks and records a verification-code request for the given IP.
	// It returns true when the request is within the allowed quota and
	// false when the quota has been exhausted.
	Allow(ip string) (bool, error)
}

var vcLimiter vcRateLimiter

// InitVCRateLimiter sets the rate limiter used for verification-code requests.
func InitVCRateLimiter(l vcRateLimiter) {
	vcLimiter = l
}

// checkVCRateLimit enforces IP-based rate limiting before a verification code
// is generated and sent. When vcLimiter is not configured (nil) the check is
// skipped, preserving backward compatibility. When the client IP cannot be
// determined the request is rejected. When the limiter returns an error or
// the quota is exhausted the request is rejected with errTooManyRequest.
func (ctl *baseController) checkVCRateLimit() *failedApiResult {
	if vcLimiter == nil {
		return nil
	}

	ip, fr := ctl.getRemoteAddr()
	if fr != nil {
		return fr
	}

	allowed, err := vcLimiter.Allow(ip)
	if err != nil {
		log.ErrorContext(ctl.Ctx.Request.Context(), "failed to check vc rate limit", "error", err)
		return newFailedApiResult(400, errTooManyRequest, fmt.Errorf("verification code is busy"))
	}

	if !allowed {
		return newFailedApiResult(400, errTooManyRequest, fmt.Errorf("verification code is busy"))
	}

	return nil
}
