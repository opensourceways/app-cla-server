package localcla

import "github.com/opensourceways/app-cla-server/signing/domain"

type LocalCLA interface {
	Remove(string) error
	AddCLA(linkId string, cla *domain.CLA) (string, error)
	LocalPath(*domain.CLAIndex) string
}
