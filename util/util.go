package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func emailToKey(email string) string {
	return strings.ReplaceAll(email, ".", "_")
}

func EmailSuffixToKey(email string) string {
	return emailToKey(strings.Split(email, "@")[1])
}

func CorporCLAPDFFile(out, claOrgID, email, other string) string {
	f := fmt.Sprintf("%s_%s%s.pdf", claOrgID, EmailSuffixToKey(email), other)
	return filepath.Join(out, f)
}

func OrgSignaturePDFFILE(out, claOrgID string) string {
	return filepath.Join(out, fmt.Sprintf("%s.pdf", claOrgID))
}

func IsFileNotExist(file string) bool {
	_, err := os.Stat(file)
	return os.IsNotExist(err)
}

// CopyBetweenStructs copy between two structs. Note: if some elements
// of 'to' are set tag of `json:"-'`, these elements will not be copied
// and should copy them manually.
func CopyBetweenStructs(from, to interface{}) error {
	d, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, to)
}
