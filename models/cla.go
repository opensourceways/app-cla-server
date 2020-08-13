package models

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

func (this *CLA) Create() error {
	v, err := db.CreateCLA(*this)
	if err == nil {
		this.ID = v
	}

	return err
}

func (this *CLA) Get() error {
	v, err := db.GetCLA(this.ID)
	if err == nil {
		*this = v
	}
	return err
}

func (this *CLA) Delete() error {
	return db.DeleteCLA(this.ID)
}

type CLAListOptions struct {
	Submitter string `json:"submitter" required:"true"`
	Name      string `json:"name,omitempty"`
	Language  string `json:"language,omitempty"`
	ApplyTo   string `json:"apply_to,omitempty"`
}

func (this CLAListOptions) Get() ([]CLA, error) {
	return db.ListCLA(this)
}
