package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type CLAInfo = dbmodels.CLAInfo

type CLAField = dbmodels.Field

type CLACreateOpt = struct {
	URL      string     `json:"url"`
	Type     string     `json:"type"`
	Fields   []CLAField `json:"fields"`
	Language string     `json:"language"`
}
