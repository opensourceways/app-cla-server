package oauth2

import (
	"io/ioutil"

	"github.com/huaweicloud/golangsdk"
	"sigs.k8s.io/yaml"
)

func loadFromYaml(path string, result interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(b, result); err != nil {
		return err
	}

	_, err = golangsdk.BuildRequestBody(result, "")
	return err
}
