package claservice

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/localcla"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewCLAService(
	repo repository.Link,
	local localcla.LocalCLA,
) CLAService {
	return &claService{
		repo:  repo,
		local: local,
	}
}

type CLAService interface {
	Add(link *domain.Link, cla *domain.CLA) error
	CLALocalFilePath(*domain.CLAIndex) string
	AddLink(link *domain.Link) error
}

type claService struct {
	repo  repository.Link
	local localcla.LocalCLA
}

func (s *claService) Add(link *domain.Link, cla *domain.CLA) error {
	if err := link.AddCLA(cla); err != nil {
		return err
	}

	p, err := s.local.AddCLA(link.Id, cla)
	if err != nil {
		return err
	}

	if err = s.repo.AddCLA(link, cla); err != nil {
		if err1 := s.local.Remove(p); err1 != nil {
			logs.Error("remove local file, err:%s", err1.Error())
		}
	}

	return err
}

func (s *claService) CLALocalFilePath(index *domain.CLAIndex) string {
	return s.local.LocalPath(index)
}

func (s *claService) AddLink(link *domain.Link) error {
	linkId := genLinkID(link)
	link.Id = linkId

	tempFiles := []string{}
	clean := func() {
		for _, p := range tempFiles {
			if err1 := s.local.Remove(p); err1 != nil {
				logs.Error("remove temp file, err:%s", err1.Error())
			}
		}
	}

	for i := range link.CLAs {
		item := &link.CLAs[i]

		p, err := s.local.AddCLA(linkId, item)
		if err != nil {
			clean()

			return err
		}

		tempFiles = append(tempFiles, p)
	}

	if err := s.repo.Add(link); err != nil {
		if commonRepo.IsErrorDuplicateCreating(err) {
			err = domain.NewDomainError(domain.ErrorCodeLinkExists)
		}

		clean()

		return err
	}

	return nil
}

func genLinkID(v *domain.Link) string {
	org := &v.Org

	return fmt.Sprintf("%s_%s-%d", org.Platform, org.Org, time.Now().UnixNano())
}
