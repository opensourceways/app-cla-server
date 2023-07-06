package dp

import "errors"

type CLAFieldType interface {
	CLAFieldType() string
}

type claFieldType string

func (v claFieldType) CLAFieldType() string {
	return string(v)
}

func NewIndividualCLAFieldType(v string) (CLAFieldType, error) {
	if v == "" || !config.isValidIndividualCLAField(v) {
		return nil, errors.New("invalid cla field type")
	}

	return claFieldType(v), nil
}

func NewCorpCLAFieldType(v string) (CLAFieldType, error) {
	if v == "" || !config.isValidCorpCLAField(v) {
		return nil, errors.New("invalid cla field type")
	}

	return claFieldType(v), nil
}
