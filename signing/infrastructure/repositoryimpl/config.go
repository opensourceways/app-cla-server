package repositoryimpl

type Config struct {
	Collections Collections `json:"collections" required:"true"`
}

type Collections struct {
	CorpSigning string `json:"corp_signing" required:"true"`
}
