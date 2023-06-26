package repositoryimpl

type Config struct {
	Collections Collections `json:"collections" required:"true"`
}

type Collections struct {
	User        string `json:"user"          required:"true"`
	CorpSigning string `json:"corp_signing"  required:"true"`
}
