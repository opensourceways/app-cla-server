package claservice

import (
	"sync"
	"time"

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
	lastUpdateTime := make(map[string]time.Time, len(links))
	for i := range links {
		v := &links[i]
		cache[v.Id] = v.Clas
		if v.LastUpdateTime > 0 {
			lastUpdateTime[v.Id] = time.Unix(v.LastUpdateTime, 0)
		}
	}

	return &linkCache{
		cache:          cache,
		lastUpdateTime: lastUpdateTime,
	}, nil
}

type linkCache struct {
	mutex          sync.RWMutex
	cache          map[string][]domain.CLA
	lastUpdateTime map[string]time.Time
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

	lc.lastUpdateTime[linkId] = time.Now()

	lc.mutex.Unlock()
}

func (lc *linkCache) removeLink(linkId string) {
	lc.mutex.Lock()
	delete(lc.cache, linkId)
	delete(lc.lastUpdateTime, linkId)
	lc.mutex.Unlock()
}

func (lc *linkCache) removeCLA(linkId, claId string) {
	lc.mutex.Lock()
	clas, ok := lc.cache[linkId]
	if !ok {
		return
	}

	n := len(clas) - 1
	for i := range clas {
		if clas[i].Id == claId {
			if i != n {
				clas[i] = clas[n]
			}
			lc.cache[linkId] = clas[:n]
			break
		}
	}

	lc.mutex.Unlock()
}

func (lc *linkCache) getLastUpdateTime(linkId string) time.Time {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	if t, ok := lc.lastUpdateTime[linkId]; ok {
		return t
	}
	return time.Time{}
}
