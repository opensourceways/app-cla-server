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

	cache := make(map[string][]domain.CLA, len(links))
	for i := range links {
		v := &links[i]
		cache[v.Id] = v.Clas
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

func (lc *linkCache) update(linkId string, newCLA *domain.CLA) {
	lc.mutex.Lock()

	if clas, ok := lc.cache[linkId]; !ok {
		lc.cache[linkId] = []domain.CLA{*newCLA}
	} else {
		bingo := false
		for i := range clas {
			if v := &clas[i]; v.Type == newCLA.Type && v.Language == newCLA.Language {
				*v = *newCLA
				bingo = true
				break
			}
		}

		if !bingo {
			lc.cache[linkId] = append(clas, *newCLA)
		}
	}

	lc.mutex.Unlock()
}

func (lc *linkCache) removeLink(linkId string) {
	lc.mutex.Lock()
	delete(lc.cache, linkId)
	lc.mutex.Unlock()
}

func (lc *linkCache) removeCLA(linkId, claId string) {
	lc.mutex.Lock()
	clas, ok := lc.cache[linkId]
	if !ok {
		return
	}

	for i := range clas {
		if clas[i].Id == claId {
			clas[i] = clas[len(clas)-1]
			break
		}
	}

	lc.cache[linkId] = clas[:len(clas)-1]

	lc.mutex.Unlock()
}
