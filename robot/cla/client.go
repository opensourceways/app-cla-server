package cla

import "fmt"

type PRInfo struct {
	Org    string
	Repo   string
	Number int
}

func (p PRInfo) String() string {
	return fmt.Sprintf("%s/%s:%d", p.Org, p.Repo, p.Number)
}

type Client interface {
	AddPRLabel(pr PRInfo, label string) error
	RemovePRLabel(pr PRInfo, label string) error
	CreatePRComment(pr PRInfo, comment string) error
	DeletePRComment(pr PRInfo, needDelete func(string) bool) error
	// return map is commit sha : commit message
	GetUnsignedCommits(pr PRInfo, commiterAsAuthor bool, isSigned func(string) bool) (map[string]string, error)
}
