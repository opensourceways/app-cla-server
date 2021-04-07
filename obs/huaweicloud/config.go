package huaweicloud

import (
	"github.com/opensourceways/app-cla-server/util"
)

type config struct {
	AccessKey           string `json:"access_key" required:"true"`
	SecretKey           string `json:"secret_key" required:"true"`
	Endpoint            string `json:"endpoint" required:"true"`
	ObjectEncryptionKey string `json:"object_encryption_key"`
}

func loadConfig(path string) (*config, error) {
	v := &config{}
	if err := util.LoadFromYaml(path, v); err != nil {
		return nil, err
	}

	return v, nil
}
