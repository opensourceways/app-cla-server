package app

import (
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

func NewCorpSigningService(
	repo repository.CorpSigning,
	vc vcservice.VCService,
	interval time.Duration,
	linkRepo repository.Link,
) *corpSigningService {
	return &corpSigningService{
		repo:     repo,
		vc:       verificationCodeService{vc},
		interval: interval,
		linkRepo: linkRepo,
	}
}

type CorpSigningService interface {
	Verify(cmd *CmdToCreateVerificationCode) (string, error)
	Sign(cmd *CmdToSignCorpCLA) error
	Remove(userId, csId string) error
	Get(userId, csId string, email dp.EmailAddr) (string, CorpSigningInfoDTO, error)
	List(userId, linkId string) ([]CorpSigningDTO, error)
	FindCorpSummary(cmd *CmdToFindCorpSummary) ([]CorpSummaryDTO, error)
}

type corpSigningService struct {
	vc       verificationCodeService
	repo     repository.CorpSigning
	interval time.Duration
	linkRepo repository.Link
}

func (s *corpSigningService) Verify(cmd *CmdToCreateVerificationCode) (string, error) {
	return s.vc.newCodeIfItCan((*cmdToCreateCodeForCorpSigning)(cmd), s.interval)
}

func (s *corpSigningService) Sign(cmd *CmdToSignCorpCLA) error {
	cmd1 := cmd.toCmd()
	if err := s.vc.validate(&cmd1, cmd.VerificationCode); err != nil {
		return err
	}

	v := cmd.toCorpSigning()

	err := s.repo.Add(&v)
	if err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			return domain.NewDomainError(domain.ErrorCodeCorpSigningReSigning)
		}
	}

	return err
}

func (s *corpSigningService) Remove(userId, csId string) error {
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

	if _, err := checkIfCommunityManager(userId, cs.Link.Id, s.linkRepo); err != nil {
		return err
	}

	return s.repo.Remove(&cs)
}

func (s *corpSigningService) Get(userId, csId string, email dp.EmailAddr) (linkId string, dto CorpSigningInfoDTO, err error) {
	item, err := s.repo.Find(csId)
	if err != nil {
		return
	}

	linkId = item.Link.Id

	if !item.IsAdmin(email) {
		if _, err = checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
			return
		}
	}

	dto = CorpSigningInfoDTO{
		Date:     item.Date,
		CLAId:    item.Link.CLAId,
		Language: item.Link.Language.Language(),
		CorpName: item.Corp.Name.CorpName(),
		RepName:  item.Rep.Name.Name(),
		RepEmail: item.Rep.EmailAddr.EmailAddr(),
		AllInfo:  item.AllInfo,
	}

	return
}

func (s *corpSigningService) List(userId, linkId string) ([]CorpSigningDTO, error) {
	if _, err := checkIfCommunityManager(userId, linkId, s.linkRepo); err != nil {
		return nil, err
	}

	v, err := s.repo.FindAll(linkId)
	if err != nil || len(v) == 0 {
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

func (s *corpSigningService) FindCorpSummary(cmd *CmdToFindCorpSummary) ([]CorpSummaryDTO, error) {
	v, err := s.repo.FindCorpSummary(cmd.LinkId, cmd.EmailAddr.Domain())
	if err != nil || len(v) == 0 {
		return nil, err
	}

	r := make([]CorpSummaryDTO, 0, len(v))
	for i := range v {
		if item := &v[i]; item.HasManager {
			r = append(r, CorpSummaryDTO{
				CorpName:      item.CorpName.CorpName(),
				CorpSigningId: item.CorpSigningId,
			})
		}
	}

	return r, nil
}
