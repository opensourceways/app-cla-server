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
	lastUpdateTime := make(map[string]map[string]time.Time, len(links))
	for i := range links {
		v := &links[i]
		cache[v.Id] = v.Clas

		// 故意不再用link级别的历史字段兜底：那个字段是整条link共用的，任何一份CLA被
		// 改动都会刷新它，拿来当“某份CLA没有自己时间戳时”的兜底，等于把“改A影响B的宽限
		// 期计时”这个问题原样带回来。这里宁可对没有自己时间戳的CLA“不知道”（下游会按安全
		// 默认处理，不会误判为过期），也不用这个会被污染的共享值。
		m := make(map[string]time.Time, len(v.Clas))
		for j := range v.Clas {
			cla := &v.Clas[j]
			if cla.UpdatedAt > 0 {
				m[claCacheKey(cla.Type, cla.Language)] = time.Unix(cla.UpdatedAt, 0)
			}
		}

		if len(m) > 0 {
			lastUpdateTime[v.Id] = m
		}
	}

	return &linkCache{
		cache:          cache,
		lastUpdateTime: lastUpdateTime,
	}, nil
}

func claCacheKey(claType dp.CLAType, language dp.Language) string {
	return claType.CLAType() + "/" + language.Language()
}

type linkCache struct {
	mutex sync.RWMutex
	cache map[string][]domain.CLA
	// lastUpdateTime: linkId -> "type/language" -> 该 CLA 槽位最后一次被替换的时间。
	// 按槽位而非按 link 记录，避免更新某一语言/类型的 CLA 时误刷新同一 link 下
	// 其他 CLA（例如企业CLA）的宽限期计时。
	lastUpdateTime map[string]map[string]time.Time
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

	key := claCacheKey(newCLA.Type, newCLA.Language)
	if lc.lastUpdateTime[linkId] == nil {
		lc.lastUpdateTime[linkId] = map[string]time.Time{}
	}
	lc.lastUpdateTime[linkId][key] = time.Now()

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

func (lc *linkCache) getLastUpdateTime(linkId string, claType dp.CLAType, language dp.Language) time.Time {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	if m, ok := lc.lastUpdateTime[linkId]; ok {
		if t, ok := m[claCacheKey(claType, language)]; ok {
			return t
		}
	}
	return time.Time{}
}
