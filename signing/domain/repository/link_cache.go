package repository

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

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

	fmt.Println(linkCache)

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

func ClaIdMatchCheck(info domain.LinkInfo, claType dp.CLAType) bool {
	fmt.Println(info, claType.CLAType())
	claId, ok := linkCache[info.Id][claType.CLAType()][info.Language.Language()]
	if !ok {
		return false
	}

	return info.CLAId == claId
}
