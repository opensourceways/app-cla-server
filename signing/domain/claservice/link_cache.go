package claservice

import (
	"sync"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func initLink(linkRepo repository.Link) (*linkCache, error) {
	links, err := linkRepo.ListAll()
	if err != nil {
		return &linkCache{}, err
	}

	cache := make(map[string][]domain.CLA)
	for _, v := range links {
		for _, c := range v.Clas {
			cache[v.Id] = append(cache[v.Id], c)
		}
	}

	return &linkCache{
		cache: cache,
	}, nil
}

type linkCache struct {
	mutex sync.RWMutex
	cache map[string][]domain.CLA
}

func (lc *linkCache) contains(linkId, claId string) bool {
	lc.mutex.RLock()
	clas, ok := lc.cache[linkId]
	lc.mutex.RUnlock()

	if !ok {
		return false
	}

	for _, v := range clas {
		if v.Id == claId {
			return true
		}
	}

	return false
}

func (lc *linkCache) getClaId(linkId string, claType dp.CLAType, language dp.Language) string {
	lc.mutex.RLock()
	clas, ok := lc.cache[linkId]
	lc.mutex.RUnlock()

	if !ok {
		return ""
	}

	for _, v := range clas {
		if v.Type == claType && v.Language == language {
			return v.Id
		}
	}

	return ""
}

func (lc *linkCache) update(linkId, claId, newClaId string, newUrl dp.URL) error {
	lc.mutex.Lock()
	clas, ok := lc.cache[linkId]
	if !ok {
		return domain.NewDomainError(domain.ErrorCodeLinkNotExists)
	}

	for k, v := range clas {
		if v.Id == claId {
			v.Id = newClaId
			v.URL = newUrl
			clas[k] = v
			break
		}
	}

	lc.cache[linkId] = clas

	lc.mutex.Unlock()

	return nil
}
