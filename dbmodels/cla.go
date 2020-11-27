package dbmodels

type CLA struct {
	Text         []byte `json:"text"`
	OrgSignature []byte `json:"org_signature"`

	CLAInfo
}

type CLAInfo struct {
	URL      string  `json:"url"`
	Language string  `json:"language"`
	Fields   []Field `json:"fields"`
}

type Field struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type CLAListOptions struct {
	Submitter string `json:"submitter"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	ApplyTo   string `json:"apply_to"`
}
