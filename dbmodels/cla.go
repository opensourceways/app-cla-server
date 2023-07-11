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

type CLAData struct {
	URL      string  `json:"url"`
	Language string  `json:"language"`
	Fields   []Field `json:"fields"`
}

type CLADetail struct {
	CLAData

	CLAId string `json:"cla_id"`
}

type CLAInfo struct {
	CLAId   string
	CLAFile string
	CLALang string
	Fields  []Field
}
