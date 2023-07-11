package models

type CLAInfo struct {
	CLAId   string
	CLAFile string
	CLALang string
	Fields  []CLAField
}

type CLACreateOpt = struct {
	URL      string     `json:"url"`
	Type     string     `json:"type"`
	Fields   []CLAField `json:"fields"`
	Language string     `json:"language"`
}

type CLAField struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type CLAData struct {
	URL      string     `json:"url"`
	Language string     `json:"language"`
	Fields   []CLAField `json:"fields"`
}

type CLADetail struct {
	CLAData

	CLAId string `json:"cla_id"`
}
