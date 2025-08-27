package claservice

import (
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/localcla"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/watch"
	"github.com/opensourceways/app-cla-server/util"
)

func NewCLAService(
	repo repository.Link,
	local localcla.LocalCLA,
	message message.Message,
) (CLAService, error) {
	cache, err := initLink(repo)
	if err != nil {
		return nil, err
	}

	return &claService{
		repo:      repo,
		local:     local,
		message:   message,
		linkCache: cache,
	}, nil
}

type CLAService interface {
	Add(link *domain.Link, cla *domain.CLA) error
	Update(link *domain.Link, newCla *domain.CLA) error
	CLALocalFilePath(*domain.CLAIndex) string
	DiffCLALocalFilePath(index *domain.CLAIndex, signedClaId string) (string, error)
	AddLink(link *domain.Link) error
	ContainsCla(linkId, claId string) bool
	GetClaId(linkId string, claType dp.CLAType, language dp.Language) string
	RemoveLink(linkId string)
	RemoveCLA(linkId, claId string)
}

type claService struct {
	linkCache *linkCache
	lock      sync.Mutex
	repo      repository.Link
	local     localcla.LocalCLA
	message   message.Message
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
	} else {
		s.linkCache.update(link.Id, cla)
	}

	return err
}

func (s *claService) Update(link *domain.Link, newCla *domain.CLA) error {
	if err := link.UpdateCLA(newCla); err != nil {
		return err
	}

	oldCLA := link.GetCLA(newCla.Type, newCla.Language)
	if oldCLA == nil {
		return domain.NewNotFoundDomainError(domain.ErrorCodeCLANotExists)
	}

	p, err := s.local.AddCLA(link.Id, newCla)
	if err != nil {
		return err
	}

	if err = s.repo.UpdateCLA(link, newCla); err != nil {
		if err1 := s.local.Remove(p); err1 != nil {
			logs.Error("remove local file, err:%s", err1.Error())
		}
	} else {
		s.linkCache.update(link.Id, newCla)

		s.message.SendCLAUpdatedEvent(message.CLAUpdatedMsg{
			LinkId:   link.Id,
			OldCLAId: oldCLA.Id,
			NewCLAId: newCla.Id,
		})
	}

	return err
}

func (s *claService) CLALocalFilePath(index *domain.CLAIndex) string {
	return s.local.LocalPath(index)
}

func (s *claService) DiffCLALocalFilePath(index *domain.CLAIndex, signedClaId string) (string, error) {
	file := s.local.LocalPathOfDiff(index, signedClaId)
	if !util.IsFileNotExist(file) {
		return file, nil
	}

	watch.SendCLAUpdatedEvent(message.CLAUpdatedMsg{
		LinkId:   index.LinkId,
		OldCLAId: signedClaId,
		NewCLAId: index.CLAId,
	})

	time.Sleep(time.Second)

	if util.IsFileNotExist(file) {
		return "", domain.NewDomainError(domain.ErrorCodeCLANotExists)
	}

	return file, nil
}

func (s *claService) AddLink(link *domain.Link) error {
	s.lock.Lock()
	linkId := s.repo.NewLinkId()
	s.lock.Unlock()

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
	} else {
		for i := range link.CLAs {
			item := &link.CLAs[i]
			s.linkCache.update(link.Id, item)
		}
	}

	return nil
}

func (s *claService) ContainsCla(linkId, claId string) bool {
	return s.linkCache.contains(linkId, claId)
}

func (s *claService) GetClaId(linkId string, claType dp.CLAType, language dp.Language) string {
	return s.linkCache.getClaId(linkId, claType, language)
}

func (s *claService) RemoveLink(linkId string) {
	s.linkCache.removeLink(linkId)
}

func (s *claService) RemoveCLA(linkId, claId string) {
	s.linkCache.removeCLA(linkId, claId)
}
