package dp

import "errors"

var (
	CLATypeCorp       = claType("corporation")
	CLATypeIndividual = claType("individual")
)

type CLAType interface {
	CLAType() string
}

type claType string

func (v claType) CLAType() string {
	return string(v)
}

func NewCLAType(v string) (CLAType, error) {
	if v == CLATypeCorp.CLAType() {
		return CLATypeCorp, nil
	}

	if v == CLATypeIndividual.CLAType() {
		return CLATypeIndividual, nil
	}

	return nil, errors.New("invalid cla type")
}

func CreateCLAType(v string) CLAType {
	return claType(v)
}

func IsCLATypeIndividual(v CLAType) bool {
	return v != nil && v.CLAType() == CLATypeIndividual.CLAType()
}
