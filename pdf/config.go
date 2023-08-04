package pdf

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/util"
)

type Config struct {
	PDFOutDir string `json:"pdf_out_dir" required:"true"`
	PythonBin string `json:"python_bin"  required:"true"`
}

func (cfg *Config) Validate() error {
	if util.IsFileNotExist(cfg.PythonBin) {
		return fmt.Errorf("the file:%s is not exist", cfg.PythonBin)
	}

	if !util.IsNotDir(cfg.PDFOutDir) {
		return nil
	}

	return util.Mkdir(cfg.PDFOutDir)
}
