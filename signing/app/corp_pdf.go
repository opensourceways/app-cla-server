package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewCorpPDFService(
	repo repository.CorpSigning,
	linkRepo repository.Link,
) CorpPDFService {
	return &corpPDFService{
		repo:     repo,
		linkRepo: linkRepo,
	}
}

type CmdToUploadCorpSigningPDF struct {
	UserId string
	CSId   string
	PDF    []byte
}

type CorpPDFService interface {
	Upload(cmd *CmdToUploadCorpSigningPDF) error
	Download(userId, csId string, userEmail dp.EmailAddr) ([]byte, error)
}

type corpPDFService struct {
	repo     repository.CorpSigning
	linkRepo repository.Link
}

func (s *corpPDFService) Upload(cmd *CmdToUploadCorpSigningPDF) error {
	cs, err := s.repo.Find(cmd.CSId)
	if err != nil {
		return err
	}

	if _, err := checkIfCommunityManager(cmd.UserId, cs.Link.Id, s.linkRepo); err != nil {
		return err
	}

	return s.repo.SaveCorpPDF(&cs, cmd.PDF)
}

func (s *corpPDFService) Download(userId, csId string, userEmail dp.EmailAddr) ([]byte, error) {
	cs, err := s.repo.Find(csId)
	if err != nil {
		return nil, err
	}

	if cs.IsAdmin(userEmail) {
		return s.repo.FindCorpPDF(csId)
	}

	if _, err := checkIfCommunityManager(userId, cs.Link.Id, s.linkRepo); err != nil {
		return nil, err
	}

	return s.repo.FindCorpPDF(csId)
}

func checkIfCommunityManager(userId, linkId string, linkRepo repository.Link) (*domain.Link, error) {
	link, err := linkRepo.Find(linkId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			err = domain.NewDomainError(domain.ErrorCodeLinkNotExists)
		}

		return nil, err
	}

	if err := link.CanDo(userId); err != nil {
		return nil, err
	}

	return &link, nil
}
