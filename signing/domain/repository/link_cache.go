package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

var linkCache map[string]map[string]map[string]string

func InitLink(linkRepo Link) error {
	linkCache = make(map[string]map[string]map[string]string)
	links, err := linkRepo.FindAll("")
	if err != nil {
		return err
	}

	for _, link := range links {
		linkCache[link.Id] = claCache(link.CLAs)
	}

	return nil
}

func claCache(clas []domain.CLA) map[string]map[string]string {
	cache := make(map[string]map[string]string)
	for _, cla := range clas {
		if _, ok := cache[cla.Type.CLAType()]; !ok {
			cache[cla.Type.CLAType()] = make(map[string]string)
		}

		cache[cla.Type.CLAType()][cla.Language.Language()] = cla.Id
	}

	return cache
}
