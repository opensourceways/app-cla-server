package util

import (
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
