package models

type CLA struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name" required:"true"`
	Text     string `json:"text" required:"true"`
	Language string `json:"language" required:"true"`
}

func (c CLA) Create() (CLA, error) {
	return db.CreateCLA(c)
}
