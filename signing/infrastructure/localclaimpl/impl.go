package localclaimpl

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/util"
)

func NewLocalCLAImpl(cfg *Config) *localCLAImpl {
	return &localCLAImpl{dir: cfg.Dir}
}

type localCLAImpl struct {
	dir string
}

func (impl *localCLAImpl) Remove(p string) error {
	return os.Remove(p)
}

func (impl *localCLAImpl) AddCLA(linkId string, cla *domain.CLA) (string, error) {
	p := impl.localPath(linkId, cla.Id)

	err := ioutil.WriteFile(p, cla.Text, 0644)

	return p, err
}

func (impl *localCLAImpl) LocalPath(index *domain.CLAIndex) string {
	return impl.localPath(index.LinkId, index.CLAId)
}

func (impl *localCLAImpl) localPath(linkId, claId string) string {
	return filepath.Join(impl.dir, linkId, claId+".pdf")
}

// config
type Config struct {
	Dir string `json:"dir" required:"true"`
}

func (cfg *Config) Validate() error {
	if util.IsNotDir(cfg.Dir) {
		if err := os.Mkdir(cfg.Dir, 0644); err != nil {
			return err
		}
	}

	return nil
}
