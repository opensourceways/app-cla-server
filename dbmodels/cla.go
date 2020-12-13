package dbmodels

const (
	ApplyToCorporation = "corporation"
	ApplyToIndividual  = "individual"
)

type CLA struct {
	Text         []byte `json:"text"`
	OrgSignature []byte `json:"org_signature"`

	CLAData
}

type CLAData struct {
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

type CLAInfo struct {
	CLALang          string
	CLAHash          string
	OrgSignatureHash string
	Fields           []Field
}
