package obs

import (
	"fmt"

	appConf "github.com/opensourceways/app-cla-server/config"
)

type OBS interface {
	Initialize(string, string) error
	WriteObject(path string, data []byte) error
	ReadObject(path, localPath string) Error
	HasObject(string) (bool, error)
	ListObject(pathPrefix string) ([]string, error)
}

var instances = map[string]OBS{}

func Register(plugin string, i OBS) {
	instances[plugin] = i
}

func Initialize(info appConf.OBS) (OBS, error) {
	i, ok := instances[info.Name]
	if !ok {
		return nil, fmt.Errorf("no such obs instance of %s", info.Name)
	}

	return i, i.Initialize(info.CredentialFile, info.Bucket)
}

type Error interface {
	Error() string
	IsObjectNotFound() bool
}