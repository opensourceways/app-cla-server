package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/huaweicloud/golangsdk"
	"sigs.k8s.io/yaml"
)

func emailToKey(email string) string {
	return strings.ReplaceAll(email, ".", "_")
}

func EmailSuffixToKey(email string) string {
	return emailToKey(strings.Split(email, "@")[1])
}

func EmailSuffix(email string) string {
	v := strings.Split(email, "@")
	if len(v) == 2 {
		return v[1]
	}
	return email
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

func IsNotDir(dir string) bool {
	v, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return !v.IsDir()
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
