package domain

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

var linkCache map[string]map[string]map[string]string

type ClaIdByLanguage struct {
	Chinese string `json:"chinese"`
	English string `json:"english"`
}

func InitLink(linkRepo repository.Link) error {
	linkCache = make(map[string]map[string]map[string]string)
	links, err := linkRepo.FindAll("")
	if err != nil {
		return err
	}

	for _, link := range links {
		linkCache[link.Id] = getTypeCache(link.CLAs)
	}

	fmt.Println(linkCache)

	return nil
}

func getTypeCache(clas []CLA) map[string]map[string]string {
	typeCache := make(map[string]map[string]string)
	for _, cla := range clas {
		if _, ok := typeCache[cla.Type.CLAType()]; !ok {
			typeCache[cla.Type.CLAType()] = make(map[string]string)
		}

		typeCache[cla.Type.CLAType()][cla.Language.Language()] = cla.Id
	}

	return typeCache
}
