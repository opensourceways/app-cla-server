package app

import "github.com/opensourceways/app-cla-server/signing/domain/repository"

func NewCorpPDFService(
	repo repository.CorpSigning,
) CorpPDFService {
	return &corpPDFService{
		repo: repo,
	}
}

type CorpPDFService interface {
	Upload(csId string, pdf []byte) error
	Download(csId string) ([]byte, error)
}

type corpPDFService struct {
	repo repository.CorpSigning
}

func (s *corpPDFService) Upload(csId string, pdf []byte) error {
	cs, err := s.repo.Find(csId)
	if err != nil {
		return err
	}

	return s.repo.SaveCorpPDF(&cs, pdf)
}

func (s *corpPDFService) Download(csId string) ([]byte, error) {
	return s.repo.FindCorpPDF(csId)
}
