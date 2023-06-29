package app

import (
	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewCorpSigningService(repo repository.CorpSigning) *corpSigningService {
	return &corpSigningService{repo}
}

type CorpSigningService interface {
	Sign(cmd *CmdToSignCorpCLA) error
	Remove(csId string) error
	Get(csId string) (CorpSigningInfoDTO, error)
	List(linkId string) ([]CorpSigningDTO, error)
}

type corpSigningService struct {
	repo repository.CorpSigning
}

func (s *corpSigningService) Sign(cmd *CmdToSignCorpCLA) error {
	v := cmd.toCorpSigning()

	err := s.repo.Add(&v)
	if err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			return domain.NewDomainError(domain.ErrorCodeCorpSigningReSigning)
		}
	}

	return err
}

func (s *corpSigningService) Remove(csId string) error {
	cs, err := s.repo.Find(csId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil
		}

		return err
	}

	if err := cs.CanRemove(); err != nil {
		return err
	}

	return s.repo.Remove(&cs)
}

func (s *corpSigningService) Get(csId string) (CorpSigningInfoDTO, error) {
	item, err := s.repo.Find(csId)
	if err != nil {
		return CorpSigningInfoDTO{}, err
	}

	return CorpSigningInfoDTO{
		Date:     item.Date,
		Language: item.Link.Language.Language(),
		CorpName: item.Corp.Name.CorpName(),
		RepName:  item.Rep.Name.Name(),
		RepEmail: item.Rep.EmailAddr.EmailAddr(),
		AllInfo:  item.AllInfo,
	}, nil
}

func (s *corpSigningService) List(linkId string) ([]CorpSigningDTO, error) {
	v, err := s.repo.FindAll(linkId)
	if err != nil {
		return nil, err
	}

	dtos := make([]CorpSigningDTO, len(v))

	for i := range v {
		item := &v[i]
		dtos[i] = CorpSigningDTO{
			Id:             item.Id,
			Date:           item.Date,
			Language:       item.Link.Language.Language(),
			CorpName:       item.Corp.Name.CorpName(),
			RepName:        item.Rep.Name.Name(),
			RepEmail:       item.Rep.EmailAddr.EmailAddr(),
			HasAdminAdded:  !item.Admin.IsEmpty(),
			HasPDFUploaded: item.HasPDF,
		}
	}

	return dtos, nil
}
