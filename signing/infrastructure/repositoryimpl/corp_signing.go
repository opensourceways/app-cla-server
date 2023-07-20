package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func NewCorpSigning(dao dao) *corpSigning {
	return &corpSigning{
		dao: dao,
	}
}

type corpSigning struct {
	dao dao
}

func (impl *corpSigning) toCorpSigningIndex(corpSigningId string) (bson.M, error) {
	return impl.dao.DocIdFilter(corpSigningId)
}

func (impl *corpSigning) Add(v *domain.CorpSigning) error {
	do := toCorpSigningDO(v)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0
	doc[fieldDeleted] = bson.A{}
	doc[fieldManagers] = bson.A{}
	doc[fieldEmployees] = bson.A{}

	docFilter := linkIdFilter(v.Link.Id)
	docFilter[mongodbCmdOr] = bson.A{
		bson.M{childField(fieldRep, fieldEmail): v.Rep.EmailAddr.EmailAddr()},
		bson.M{
			childField(fieldCorp, fieldName):   v.Corp.Name.CorpName(),
			childField(fieldCorp, fieldDomain): v.Corp.PrimaryEmailDomain,
		},
	}

	_, err = impl.dao.InsertDocIfNotExists(docFilter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *corpSigning) Remove(cs *domain.CorpSigning) error {
	filter, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}
	filter[fieldVersion] = cs.Version

	if err = impl.dao.DeleteDoc(filter); err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *corpSigning) Find(index string) (cs domain.CorpSigning, err error) {
	filter, err := impl.toCorpSigningIndex(index)
	if err != nil {
		return
	}

	project := bson.M{
		fieldPDF:     0,
		fieldDeleted: 0,
	}

	var do corpSigningDO

	if err = impl.dao.GetDoc(filter, project, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}
	} else {
		err = do.toCorpSigning(&cs)
	}

	return
}

func (impl *corpSigning) FindCorpSummary(linkId, domain string) ([]repository.CorpSummary, error) {
	filter := linkIdFilter(linkId)
	filter[childField(fieldCorp, fieldDomains)] = bson.M{mongodbCmdIn: bson.A{domain}}

	var dos []corpSigningDO

	if err := impl.dao.GetDocs(filter, bson.M{fieldCorp: 1}, &dos); err != nil {
		return nil, err
	}

	v := make([]repository.CorpSummary, len(dos))
	for i := range dos {
		item := &dos[i]

		corp, err := item.Corp.toCorp()
		if err != nil {
			return nil, err
		}

		v[i] = repository.CorpSummary{
			CorpSigningId: item.index(),
			CorpName:      corp.Name,
		}
	}

	return v, nil
}

func (impl *corpSigning) FindAll(linkId string) ([]repository.CorpSigningSummary, error) {
	filter := linkIdFilter(linkId)

	project := bson.M{
		fieldDate:   1,
		fieldLang:   1,
		fieldRep:    1,
		fieldCorp:   1,
		fieldAdmin:  1,
		fieldLinkId: 1,
		fieldHasPDF: 1,
	}

	var dos []corpSigningDO

	if err := impl.dao.GetDocs(filter, project, &dos); err != nil {
		return nil, err
	}

	v := make([]repository.CorpSigningSummary, len(dos))
	for i := range dos {
		if err := dos[i].toCorpSigningSummary(&v[i]); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (impl *corpSigning) HasSignedLink(linkId string) (bool, error) {
	filter := linkIdFilter(linkId)

	var do corpSigningDO

	if err := impl.dao.GetDoc(filter, bson.M{fieldLinkId: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil

}

func (impl *corpSigning) HasSignedCLA(index *domain.CLAIndex, t dp.CLAType) (bool, error) {
	if dp.IsCLATypeIndividual(t) {
		return impl.hasSignedEmployeeCLA(index)
	}

	return impl.hasSignedCorpCLA(index)
}

func (impl *corpSigning) hasSignedCorpCLA(index *domain.CLAIndex) (bool, error) {
	filter := linkIdFilter(index.LinkId)
	filter[fieldCLAId] = index.CLAId

	var do corpSigningDO

	if err := impl.dao.GetDoc(filter, bson.M{fieldLinkId: 1}, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
