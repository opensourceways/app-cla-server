package dp

import "errors"

var (
	LinkTypeCLA = linkType("cla")
	LinkTypeDCO = linkType("dco")
)

type LinkType interface {
	LinkType() string
}

type linkType string

func (v linkType) LinkType() string {
	return string(v)
}

func NewLinkType(v string) (LinkType, error) {
	if v == LinkTypeCLA.LinkType() {
		return LinkTypeCLA, nil
	}

	if v == LinkTypeDCO.LinkType() {
		return LinkTypeDCO, nil
	}

	return nil, errors.New("invalid link type")
}

func IsLinkTypeCLA(v LinkType) bool {
	return v != nil && v.LinkType() == LinkTypeCLA.LinkType()
}
