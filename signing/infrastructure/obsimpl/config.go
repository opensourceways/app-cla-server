package obsimpl

type Config struct {
	Endpoint   string `json:"endpoint"     required:"true"`
	Bucket     string `json:"bucket"       required:"true"`
	AccessKey  string `json:"access_key"   required:"true"`
	SecretKey  string `json:"secret_key"   required:"true"`
	PathPrefix string `json:"path_prefix"`
}

func (cfg *Config) SetDefault() {
	if cfg.PathPrefix == "" {
		cfg.PathPrefix = "cla-pdfs"
	}
}

func (cfg *Config) Validate() error {
	cfg.SetDefault()

	return nil
}

func (cfg *Config) objectKey(corpSigningId string) string {
	return cfg.PathPrefix + "/corp-signing/" + corpSigningId + ".pdf"
}
