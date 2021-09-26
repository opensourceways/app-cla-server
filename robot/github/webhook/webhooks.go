/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"errors"
	"net/http"
)

// ValidateWebhook ensures that the provided request conforms to the
// format of a GitHub webhook and the payload can be validated with
// the provided hmac secret. It returns the event type, the event guid,
// the payload of the request, whether the webhook is valid or not,
// and finally the resultant HTTP status code
func ValidateWebhook(getHeader func(string) string, payload []byte, tokenGenerator func() []byte) (eventType string, eventGUID string, code int, err error) {
	if eventType = getHeader("X-GitHub-Event"); eventType == "" {
		code = http.StatusBadRequest
		err = errors.New("Missing X-GitHub-Event Header")
		return
	}

	if eventGUID = getHeader("X-GitHub-Delivery"); eventGUID == "" {
		code = http.StatusBadRequest
		err = errors.New("Missing X-GitHub-Delivery Header")
		return
	}

	if contentType := getHeader("content-type"); contentType != "application/json" {
		code = http.StatusBadRequest
		err = errors.New("only accepts content-type: application/json - please reconfigure this hook on GitHub")
		return
	}

	sig := getHeader("X-Hub-Signature")
	if sig == "" {
		code = http.StatusForbidden
		err = errors.New("Missing X-Hub-Signature")
		return
	}

	// Validate the payload with our HMAC secret.
	if err = ValidatePayload(payload, sig, tokenGenerator); err != nil {
		code = http.StatusForbidden
	} else {
		code = http.StatusOK
	}
	return
}
