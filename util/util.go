package util

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"sigs.k8s.io/yaml"
)

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

func LoadFromYaml(path string, cfg interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	content := []byte(os.ExpandEnv(string(b)))

	if err := yaml.Unmarshal(content, cfg); err != nil {
		return err
	}

	_, err = BuildRequestBody(cfg, "")
	return err
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
	if data == nil {
		return ""
	}

	return fmt.Sprintf("%x", md5.Sum(data))
}

func GenFileName(fileNameParts ...string) string {
	s := filepath.Join(fileNameParts...)
	return strings.ReplaceAll(s, string(filepath.Separator), "_")
}

func CheckContentType(data []byte, t string) bool {
	s := http.DetectContentType(data)

	return strings.Contains(strings.ToLower(s), t)
}
