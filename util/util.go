package util

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/huaweicloud/golangsdk"
	"sigs.k8s.io/yaml"
)

func EmailSuffix(email string) string {
	v := strings.Split(email, "@")
	if len(v) == 2 {
		return v[1]
	}
	return email
}

func CorporCLAPDFFile(out, claOrgID, email, other string) string {
	s := strings.ReplaceAll(EmailSuffix(email), ".", "_")
	f := fmt.Sprintf("%s_%s%s.pdf", claOrgID, s, other)
	return filepath.Join(out, f)
}

func OrgSignaturePDFFILE(out, claOrgID string) string {
	return filepath.Join(out, fmt.Sprintf("%s.pdf", claOrgID))
}

func GenFilePath(dir, fileName string) string {
	return filepath.Join(dir, fileName)
}

func IsFileNotExist(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return false
	}
	return true
}

func IsNotDir(dir string) bool {
	v, err := os.Stat(dir)
	if err == nil {
		return !v.IsDir()
	}
	return true
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

func LoadFromYaml(path string, cfg interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(b, cfg); err != nil {
		return err
	}

	_, err = golangsdk.BuildRequestBody(cfg, "")
	return err
}

func NewTemplate(name, path string) (*template.Template, error) {
	txtStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to new template: read template file failed: %s", err.Error())
	}

	tmpl, err := template.New(name).Parse(string(txtStr))
	if err != nil {
		return nil, fmt.Errorf("Failed to new template: build template failed: %s", err.Error())
	}

	return tmpl, nil
}

func RenderTemplate(tmpl *template.Template, data interface{}) (string, error) {
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("Failed to execute template(%s): %s", tmpl.Name(), err.Error())
	}

	return buf.String(), nil
}

func Date() string {
	return time.Now().Format("2006-01-02")
}

func Now() int64 {
	return time.Now().Unix()
}

func Expiry(expiry int64) int64 {
	return time.Now().Add(time.Second * time.Duration(expiry)).Unix()
}

func RandStr(strSize int, randType string) string {
	var dictionary string

	switch randType {
	case "alphanum":
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "alpha":
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "number":
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)

	n := byte(len(dictionary))
	for k, v := range bytes {
		bytes[k] = dictionary[v%n]
	}
	return string(bytes)
}

func Md5sumOfFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return Md5sumOfBytes(data), nil
}

func Md5sumOfBytes(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func ProjectURL(platform, org, repo string) string {
	if repo == "" {
		return fmt.Sprintf("https://%s.com/%s", platform, org)
	}
	return fmt.Sprintf("https://%s.com/%s/%s", platform, org, repo)
}

func GenFileName(fileNameParts ...string) string {
	s := filepath.Join(fileNameParts...)
	return strings.ReplaceAll(s, string(filepath.Separator), "_")
}
