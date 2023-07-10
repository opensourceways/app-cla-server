package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewCorpEmailDomainService(
	vc vcservice.VCService,
	repo repository.CorpSigning,
) CorpEmailDomainService {
	return &corpEmailDomainService{
		repo: repo,
		vc:   verificationCodeService{vc},
	}
}

type CorpEmailDomainService interface {
	Verify(cmd *CmdToVerifyEmailDomain) (string, error)
	Add(cmd *CmdToAddEmailDomain) error
	List(string) ([]string, error)
}

type corpEmailDomainService struct {
	vc   verificationCodeService
	repo repository.CorpSigning
}

func (s *corpEmailDomainService) Verify(cmd *CmdToVerifyEmailDomain) (string, error) {
	return s.vc.newCode(cmd)
}

func (s *corpEmailDomainService) Add(cmd *CmdToAddEmailDomain) error {
	if err := s.vc.validate(cmd, cmd.VerificationCode); err != nil {
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
