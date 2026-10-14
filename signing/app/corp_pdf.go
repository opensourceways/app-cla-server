package app

import (
	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/obsimpl"
)

func NewCorpPDFService(
	repo repository.CorpSigning,
	linkRepo repository.Link,
	obs obsimpl.PDFStorage,
) CorpPDFService {
	return &corpPDFService{
		repo:     repo,
		linkRepo: linkRepo,
		obs:      obs,
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
	MigrateCorpPDFsToOBS(batchSize int) error
}

type corpPDFService struct {
	repo     repository.CorpSigning
	linkRepo repository.Link
	obs      obsimpl.PDFStorage
}

func (s *corpPDFService) Upload(cmd *CmdToUploadCorpSigningPDF) error {
	cs, err := s.repo.Find(cmd.CSId)
	if err != nil {
		return err
	}

	if _, err := checkIfCommunityManager(cmd.UserId, cs.Link.Id, s.linkRepo); err != nil {
		return err
	}

	url, err := s.obs.Upload(cmd.CSId, cmd.PDF)
	if err != nil {
		logs.Error("upload corp pdf to obs failed, csId=%s, err=%s", cmd.CSId, err.Error())
		return domain.NewDomainError(domain.ErrorCodeCorpPDFUploadFailed)
	}

	return s.repo.SaveCorpPDFURL(&cs, url)
}

func (s *corpPDFService) Download(userId, csId string, userEmail dp.EmailAddr) ([]byte, error) {
	cs, err := s.repo.Find(csId)
	if err != nil {
		return nil, err
	}

	if !cs.IsAdmin(userEmail) {
		if _, err := checkIfCommunityManager(userId, cs.Link.Id, s.linkRepo); err != nil {
			return nil, err
		}
	}

	return s.downloadPDF(csId)
}

func (s *corpPDFService) downloadPDF(csId string) ([]byte, error) {
	url, err := s.repo.FindCorpPDFURL(csId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil, domain.NewNotFoundDomainError(domain.ErrorCodeCorpPDFNotFound)
		}
		return nil, err
	}

	if url != "" {
		data, err := s.obs.Download(url)
		if err != nil {
			logs.Error("download corp pdf from obs failed, csId=%s, key=%s, err=%s", csId, url, err.Error())
			return nil, domain.NewNotFoundDomainError(domain.ErrorCodeCorpPDFNotFound)
		}

		return data, nil
	}

	pdf, err := s.repo.FindCorpPDF(csId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil, domain.NewNotFoundDomainError(domain.ErrorCodeCorpPDFNotFound)
		}
		return nil, err
	}

	if len(pdf) == 0 {
		return nil, domain.NewNotFoundDomainError(domain.ErrorCodeCorpPDFNotFound)
	}

	return pdf, nil
}

func (s *corpPDFService) MigrateCorpPDFsToOBS(batchSize int) error {
	if batchSize <= 0 {
		batchSize = 50
	}

	ids, total, err := s.repo.FindCorpSigningIDsForPDFMigration(1, 0)
	if err != nil {
		logs.Error("find corp signing ids for pdf migration failed, err=%s", err.Error())
		return err
	}

	if len(ids) == 0 {
		logs.Info("pdf migration: no documents to migrate")
		return nil
	}

	logs.Info("pdf migration: starting migration of %d documents", total)

	var migrated, failed int
	for i, id := range ids {
		if err := s.migrateOneCorpPDF(id); err != nil {
			logs.Error("migrate corp pdf failed, csId=%s, err=%s", id, err.Error())
			failed++
		} else {
			migrated++
		}

		if (i+1)%batchSize == 0 {
			logs.Info("pdf migration progress: %d/%d, migrated=%d, failed=%d", i+1, len(ids), migrated, failed)
		}
	}

	logs.Info("pdf migration completed: total=%d, migrated=%d, failed=%d", len(ids), migrated, failed)

	return nil
}

func (s *corpPDFService) migrateOneCorpPDF(csId string) error {
	pdf, err := s.repo.FindCorpPDF(csId)
	if err != nil {
		if commonRepo.IsErrorResourceNotFound(err) {
			return nil
		}
		return err
	}

	if len(pdf) == 0 {
		return nil
	}

	url, err := s.obs.Upload(csId, pdf)
	if err != nil {
		return err
	}

	return s.repo.SetCorpPDFURLAndUnsetPDF(csId, url)
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
