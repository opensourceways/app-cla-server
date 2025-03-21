package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"sigs.k8s.io/yaml"
)

var reXSS = regexp.MustCompile(`[&<>"'/]`)

func HasXSS(s string) bool {
	return reXSS.MatchString(s)
}

func StrLen(s string) int {
	return utf8.RuneCountInString(s)
}

func EmailSuffix(email string) string {
	v := strings.Split(email, "@")
	if len(v) == 2 {
		return v[1]
	}
	return email
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

func Mkdir(p string) error {
	return os.MkdirAll(p, 0770)
}

func LoadFromYaml(path string, cfg interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	content := []byte(os.ExpandEnv(string(b)))

	return yaml.Unmarshal(content, cfg)
}

func NewTemplate(name, path string) (*template.Template, error) {
	txtStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to new template: read template file failed: %s", err.Error())
	}

	tmpl, err := template.New(name).Parse(string(txtStr))
	if err != nil {
		return nil, fmt.Errorf("failed to new template: build template failed: %s", err.Error())
	}

	return tmpl, nil
}

func RenderTemplate(tmpl *template.Template, data interface{}) (bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, data)
	if err != nil {
		err = fmt.Errorf("failed to execute template(%s): %s", tmpl.Name(), err.Error())
	}

	return *buf, err
}

func Date() string {
	return time.Now().Format("2006-01-02")
}

func Time() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Now() int64 {
	return time.Now().Unix()
}

func Expiry(expiry int64) int64 {
	return time.Now().Add(time.Second * time.Duration(expiry)).Unix()
}

func CheckContentType(data []byte, t string) bool {
	s := http.DetectContentType(data)

	return strings.Contains(strings.ToLower(s), t)
}
