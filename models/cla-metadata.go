package models

type CLAMetadata struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name" required:"true"`
	Text      string `json:"text" required:"true"`
	Language  string `json:"language" required:"true"`
	Submitter string `json:"submitter" required:"true"`
}

func (this *CLAMetadata) Create() error {
	v, err := db.CreateCLAMetadata(*this)
	if err == nil {
		this.ID = v
	}

	return err
}

func (this *CLAMetadata) Get() error {
	v, err := db.GetCLAMetadata(this.ID)
	if err == nil {
		*this = v
	}
	return err
}

func (this *CLAMetadata) Delete() error {
	return db.DeleteCLAMetadata(this.ID)
}

type CLAMetadatas struct {
	BelongTo []string
}

func (this CLAMetadatas) Get() ([]CLAMetadata, error) {
	return db.ListCLAMetadata(this.BelongTo)
}
