package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewCorpEmailDomainService(
	repo repository.CorpSigning,
) CorpEmailDomainService {
	return &corpEmailDomainService{
		repo: repo,
	}
}

type CorpEmailDomainService interface {
	Verify(cmd *CmdToVerifyEmailDomain) (string, error)
	Add(cmd *CmdToAddEmailDomain) error
	List(string) ([]string, error)
}

type corpEmailDomainService struct {
	vc   vcservice.VCService
	repo repository.CorpSigning
}

func (s *corpEmailDomainService) Verify(cmd *CmdToVerifyEmailDomain) (string, error) {
	p, err := cmd.purpose()
	if err != nil {
		return "", err
	}

	return s.vc.New(p)
}

func (s *corpEmailDomainService) Add(cmd *CmdToAddEmailDomain) error {
	k, err := cmd.key()
	if err != nil {
		return err
	}

	if err := s.vc.Verify(&k); err != nil {
		return err
	}

	cs, err := s.repo.Find(cmd.CorpSigningId)
	if err != nil {
		return err
	}

	if err := cs.AddEmailDomain(cmd.EmailAddr); err != nil {
		return err
	}

	return s.repo.AddEmailDomain(&cs, cmd.EmailAddr.Domain())
}

func (s *corpEmailDomainService) List(csId string) ([]string, error) {
	return s.repo.FindEmailDomains(csId)
}
