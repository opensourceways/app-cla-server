package models

import "github.com/zengchen1024/cla-server/dbmodels"

const (
	ApplyToCorporation = "corporation"
	ApplyToIndividual  = "individual"
)

type CLA struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Text      string  `json:"text"`
	Language  string  `json:"language"`
	Submitter string  `json:"submitter"`
	ApplyTo   string  `json:"apply_to"`
	Fields    []Field `json:"fields"`
}

type Field struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

func (this *CLA) Create() error {
	p := dbmodels.CLA{}
	if err := copyBetweenStructs(this, &p); err != nil {
		return err
	}
	v, err := dbmodels.GetDB().CreateCLA(p)
	if err == nil {
		this.ID = v
	}

	return err
}

func (this *CLA) Get() error {
	v, err := dbmodels.GetDB().GetCLA(this.ID)
	if err == nil {
		return copyBetweenStructs(&v, this)
	}
	return err
}

func (this *CLA) Delete() error {
	return dbmodels.GetDB().DeleteCLA(this.ID)
}

type CLAListOptions struct {
	Submitter string `json:"submitter"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	ApplyTo   string `json:"apply_to"`
}

func (this CLAListOptions) Get() ([]dbmodels.CLA, error) {
	p := dbmodels.CLAListOptions{}
	if err := copyBetweenStructs(&this, &p); err != nil {
		return nil, err
	}
	return dbmodels.GetDB().ListCLA(p)
}
