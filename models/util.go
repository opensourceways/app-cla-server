package models

import (
	"encoding/json"
	"strings"
)

func emailToKey(email string) string {
	return strings.ReplaceAll(email, ".", "_")
}

func emailSuffixToKey(email string) string {
	return emailToKey(strings.Split(email, "@")[1])
}

// copyBetweenStructs copy between two structs. Note: if some elements
// of 'to' are set tag of `json:"-'`, these elements will not be copied
// and should copy them manually.
func copyBetweenStructs(from, to interface{}) error {
	d, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, to)
}
