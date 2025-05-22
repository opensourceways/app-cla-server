package claservice

import (
	"sync"

	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

type linkCache struct {
	mutex sync.RWMutex
	cache map[string]claIdMap
}

type claIdMap map[string]bool

func InitLink(linkRepo repository.Link) (*linkCache, error) {
	links, err := linkRepo.ListAll()
	if err != nil {
		return &linkCache{}, err
	}

	cache := make(map[string]claIdMap)
	for _, v := range links {
		idMap := make(map[string]bool)
		for _, id := range v.Clas {
			idMap[id] = true
		}

		cache[v.Id] = idMap
	}

	return &linkCache{
		cache: cache,
	}, nil
}

func (lc *linkCache) isSameVersion(linkId, claId string) bool {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	_, ok := lc.cache[linkId][claId]

	return ok
}

func (lc *linkCache) update(linkId, claId string) bool {
	lc.mutex.Lock()
	lc.mutex.Unlock()

	return false
}
