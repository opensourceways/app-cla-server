package dbmodels

type CLA struct {
	ID        string  `json:"id,omitempty"`
	Name      string  `json:"name" required:"true"`
	Text      string  `json:"text" required:"true"`
	Language  string  `json:"language" required:"true"`
	Submitter string  `json:"submitter" required:"true"`
	ApplyTo   string  `json:"apply_to" required:"true"`
	Fields    []Field `json:"fields,omitempty"`
}

type Field struct {
	ID          string `json:"id" required:"true"`
	Title       string `json:"title" required:"true"`
	Type        string `json:"type" required:"true"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required" required:"true"`
}

type CLAListOptions struct {
	Submitter string `json:"submitter" required:"true"`
	Name      string `json:"name,omitempty"`
	Language  string `json:"language,omitempty"`
	ApplyTo   string `json:"apply_to,omitempty"`
}
