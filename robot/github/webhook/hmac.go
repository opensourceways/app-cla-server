/*
Copyright 2016 The Kubernetes Authors.

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
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

// hmacSecret contains a hmac token and the time when it's created.
type hmacSecret struct {
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// hmacsForRepo contains all hmac tokens configured for a repo, org or globally.
type hmacsForRepo []hmacSecret

// ValidatePayload ensures that the request payload signature matches the key.
func ValidatePayload(payload []byte, sig string, tokenGenerator func() []byte) error {
	var event GenericEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("couldn't unmarshal the github event payload. err:%s", err.Error())
	}

	if !strings.HasPrefix(sig, "sha1=") {
		return fmt.Errorf("invalid signature")
	}
	sig = sig[5:]

	sb, err := hex.DecodeString(sig)
	if err != nil {
		return fmt.Errorf("unkown signature")
	}

	hmacs, err := extractHmacs(event.Repo.GetFullName(), tokenGenerator)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal the hmac secret. err:%s", err.Error())
	}

	// If we have a match with any valid hmac, we can validate successfully.
	for _, key := range hmacs {
		mac := hmac.New(sha1.New, key)
		if _, err := mac.Write(payload); err != nil {
			return err
		}

		expected := mac.Sum(nil)
		if hmac.Equal(sb, expected) {
			return nil
		}
	}

	return fmt.Errorf("no matched hmac")
}

// PayloadSignature returns the signature that matches the payload.
func PayloadSignature(payload []byte, key []byte) string {
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(payload)
	sum := mac.Sum(nil)
	return "sha1=" + hex.EncodeToString(sum)
}

// extractHmacs returns all *valid* HMAC tokens for given repository/organization.
// It considers only the tokens at the most specific level configured for the given repo.
// For example : if a token for repo is present and it doesn't match the repo, we will
// not try to find a match with org level token. However if no token is present for repo,
// we will try to match with org level.
func extractHmacs(repo string, tokenGenerator func() []byte) ([][]byte, error) {
	t := tokenGenerator()
	repoToTokenMap := map[string]hmacsForRepo{}

	if err := yaml.Unmarshal(t, &repoToTokenMap); err != nil {
		// To keep backward compatibility, we are going to assume that in case of error,
		// whole file is a single line hmac token.
		//TODO : Once this code has been released and file has been moved to new format,
		// we should delete this code and return error.
		return [][]byte{t}, nil
	}

	orgName := strings.Split(repo, "/")[0]

	if val, ok := repoToTokenMap[repo]; ok {
		return extractTokens(val), nil
	}
	if val, ok := repoToTokenMap[orgName]; ok {
		return extractTokens(val), nil
	}
	if val, ok := repoToTokenMap["*"]; ok {
		return extractTokens(val), nil
	}
	return nil, errors.New("invalid content in secret file, global token doesn't exist")
}

// extractTokens return tokens for any given level of tree.
func extractTokens(allTokens hmacsForRepo) [][]byte {
	validTokens := make([][]byte, len(allTokens))
	for i := range allTokens {
		validTokens[i] = []byte(allTokens[i].Value)
	}
	return validTokens
}
