package dbmodels

const (
	ApplyToCorporation = "corporation"
	ApplyToIndividual  = "individual"
)

type CLA struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Text     string  `json:"text"`
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

type CLAData struct {
	URL      string  `json:"url"`
	Language string  `json:"language"`
	Fields   []Field `json:"fields"`
}

type CLADetail struct {
	CLAData
	CLAHash string `json:"cla_hash"`
	Text    string `json:"text"`
}

type CLACreateOption struct {
	OrgSignature     *[]byte `json:"org_signature"`
	OrgSignatureHash string

	CLADetail
}

type CLAInfo struct {
	CLALang          string
	CLAHash          string
	OrgSignatureHash string
	Fields           []Field
}
