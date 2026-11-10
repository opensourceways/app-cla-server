package repositoryimpl

import (
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

// CorpSigningIndexes returns the index definitions that should exist on the
// corp_signing collection. Pass the result to mongodb.EnsureIndexes on startup.
func CorpSigningIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		// Fast lookup by link — the primary query key for all page/list queries.
		{Keys: bson.D{{Key: fieldLinkId, Value: 1}}},
		// Compound index for the adminAdded filter (link_id + admin.id).
		{Keys: bson.D{{Key: fieldLinkId, Value: 1}, {Key: "admin.id", Value: 1}}},
		// Compound index to support FindPage's $match(link_id) + $sort(date: -1),
		// so the sort can be served by the index instead of an in-memory sort.
		{Keys: bson.D{{Key: fieldLinkId, Value: 1}, {Key: fieldDate, Value: -1}}},
	}
}

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

	id, err := impl.dao.InsertDocIfNotExists(docFilter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}
	v.Id = id
	return err
}

func (impl *corpSigning) AddForMigrate(v *domain.CorpSigning) error {
	do := toCorpSigningDOForMigrate(v)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0
	doc[fieldDeleted] = bson.A{}

	docFilter := linkIdFilter(v.Link.Id)
	docFilter[mongodbCmdOr] = bson.A{
		bson.M{childField(fieldRep, fieldEmail): v.Rep.EmailAddr.EmailAddr()},
		bson.M{
			childField(fieldCorp, fieldName):   v.Corp.Name.CorpName(),
			childField(fieldCorp, fieldDomain): v.Corp.PrimaryEmailDomain,
		},
	}

	id, err := impl.dao.InsertDocIfNotExists(docFilter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}
	v.Id = id
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
		cs = do.toCorpSigning()
	}

	return
}

func (impl *corpSigning) FindCorpSummary(linkId, domain string) ([]repository.CorpSummary, error) {
	filter := linkIdFilter(linkId)
	filter[childField(fieldCorp, fieldDomains)] = bson.M{mongodbCmdIn: bson.A{domain}}
	filter[childField(fieldAdmin, fieldEmail)] = bson.M{mongodbCmdNe: ""}

	var dos []corpSigningDO

	if err := impl.dao.GetDocs(filter, bson.M{fieldCorp: 1, fieldAdmin: 1}, &dos); err != nil {
		return nil, err
	}

	v := make([]repository.CorpSummary, len(dos))
	for i := range dos {
		item := &dos[i]
		corp := item.Corp.toCorp()

		v[i] = repository.CorpSummary{
			CorpName:      corp.Name,
			CorpSigningId: item.index(),
		}
	}

	return v, nil
}

func (impl *corpSigning) FindAll(linkId string) ([]repository.CorpSigningSummary, error) {
	filter := linkIdFilter(linkId)

	project := bson.M{
		fieldDate:           1,
		fieldCLAId:          1,
		fieldLang:           1,
		fieldRep:            1,
		fieldCorp:           1,
		fieldAdmin:          1,
		fieldLinkId:         1,
		fieldHasPDF:         1,
		fieldCLANotify:      1,
		fieldClaNotifyCount: 1,
		fieldClaNotifyTime:  1,
	}

	var dos []corpSigningDO

	if err := impl.dao.GetDocs(filter, project, &dos); err != nil {
		return nil, err
	}

	v := make([]repository.CorpSigningSummary, len(dos))
	for i := range dos {
		v[i] = dos[i].toCorpSigningSummary()
	}

	return v, nil
}

func (impl *corpSigning) FindAllWithPagination(linkId string, offset, limit int) ([]repository.CorpSigningSummary, error) {
	// 参数校验
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 1000 { // 限制最大查询数量
		limit = 50
	}
	filter := linkIdFilter(linkId)

	project := bson.M{
		fieldDate:           1,
		fieldCLAId:          1,
		fieldLang:           1,
		fieldRep:            1,
		fieldCorp:           1,
		fieldAdmin:          1,
		fieldLinkId:         1,
		fieldHasPDF:         1,
		fieldCLANotify:      1,
		fieldClaNotifyCount: 1,
		fieldClaNotifyTime:  1,
	}

	var dos []corpSigningDO

	// 使用分页查询
	if err := impl.dao.GetDocsWithPagination(filter, project, offset, limit, &dos); err != nil {
		return nil, err
	}

	result := make([]repository.CorpSigningSummary, len(dos))
	for i := range dos {
		result[i] = dos[i].toCorpSigningSummary()
	}

	return result, nil
}

func (impl *corpSigning) CountByLinkId(linkId string) (int64, error) {
	filter := linkIdFilter(linkId)
	filter[fieldDeleted] = bson.M{"$size": 0}
	return impl.dao.CountDocs(filter)
}


func (impl *corpSigning) FindPage(linkId string, intPage, intPageSize int, adminAdded bool, searchQuery string) (repository.CorpSigningSummaryPage, error) {
	filter := linkIdFilter(linkId)

	var searchCond bson.M
	if searchQuery != "" {
		regex := primitive.Regex{Pattern: regexp.QuoteMeta(searchQuery), Options: "i"}
		searchCond = bson.M{"$or": []bson.M{
			{childField(fieldRep, fieldEmail): regex},
			{childField(fieldCorp, fieldName): regex},
		}}
	}

	var adminCond bson.M
	if adminAdded {
		adminCond = bson.M{"admin.id": bson.M{"$ne": ""}}
	} else {
		adminCond = bson.M{"$or": []bson.M{
			{"admin.id": ""},
			{"admin": bson.M{"$exists": false}},
		}}
	}

	if searchCond != nil && adminCond != nil {
		filter["$and"] = []bson.M{searchCond, adminCond}
	} else if searchCond != nil {
		for k, v := range searchCond {
			filter[k] = v
		}
	} else if adminCond != nil {
		for k, v := range adminCond {
			filter[k] = v
		}
	}

	project := bson.M{
		fieldDate:           1,
		fieldCLAId:          1,
		fieldLang:           1,
		fieldRep:            1,
		fieldCorp:           1,
		fieldAdmin:          1,
		fieldLinkId:         1,
		fieldHasPDF:         1,
		fieldCLANotify:      1,
		fieldClaNotifyCount: 1,
		fieldClaNotifyTime:  1,
	}

	pipeline := bson.A{
		bson.M{"$match": filter},
		bson.M{"$sort": bson.D{{Key: fieldDate, Value: -1}}},
		bson.M{"$facet": bson.M{
			"total": bson.A{
				bson.M{"$count": "n"},
			},
			"data": bson.A{
				bson.M{"$skip": int64((intPage - 1) * intPageSize)},
				bson.M{"$limit": int64(intPageSize)},
				bson.M{"$project": project},
			},
		}},
	}

	var facetResult []struct {
		Total []struct {
			N int64 `bson:"n"`
		} `bson:"total"`
		Data []corpSigningDO `bson:"data"`
	}

	if err := impl.dao.Aggregate(pipeline, &facetResult); err != nil {
		return repository.CorpSigningSummaryPage{}, err
	}

	var page repository.CorpSigningSummaryPage
	if len(facetResult) == 0 || len(facetResult[0].Total) == 0 {
		return page, nil
	}

	page.Total = facetResult[0].Total[0].N
	if page.Total == 0 {
		return page, nil
	}

	dos := facetResult[0].Data
	page.Data = make([]repository.CorpSigningSummary, len(dos))
	for i := range dos {
		page.Data[i] = dos[i].toCorpSigningSummary()
	}

	return page, nil
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

func (impl *corpSigning) UpdateClaId(cs *domain.CorpSigning) error {
	filter, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	return impl.dao.UpdateDoc(filter, bson.M{
		fieldCLAId:          cs.Link.CLAId,
		fieldCLANotify:      "",
		fieldClaNotifyCount: 0,
		fieldClaNotifyTime:  0,
	}, cs.Version)
}

func (impl *corpSigning) UpdateCLANotify(summary *repository.CorpSigningSummary) error {
	filter, err := impl.toCorpSigningIndex(summary.Id)
	if err != nil {
		return err
	}

	return impl.dao.UpdateDocsWithoutVersion(filter, bson.M{
		fieldCLANotify:      summary.CLANotify,
		fieldClaNotifyCount: summary.ClaNotifyCount,
		fieldClaNotifyTime:  summary.ClaNotifyTime,
	})
}

func (impl *corpSigning) UpdateAutoApprove(cs *domain.CorpSigning) error {
	filter, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	return impl.dao.UpdateDocsWithoutVersion(filter, bson.M{
		fieldAutoApprove: cs.AutoApproveEmployees,
	})
}

func (impl *corpSigning) Update(cs *domain.CorpSigning) error {
	filter, err := impl.toCorpSigningIndex(cs.Id)
	if err != nil {
		return err
	}

	// 构建更新文档
	updateDoc := bson.M{
		childField(fieldRep, fieldName):  cs.Rep.Name.Name(),
		childField(fieldRep, fieldEmail): cs.Rep.EmailAddr.EmailAddr(),
	}

	// 同步更新 Admin 信息
	if cs.Admin.Id != "" {
		updateDoc[childField(fieldAdmin, fieldName)] = cs.Admin.Name.Name()
		updateDoc[childField(fieldAdmin, fieldEmail)] = cs.Admin.EmailAddr.EmailAddr()
	}

	return impl.dao.UpdateDoc(filter, updateDoc, cs.Version)
}

func (impl *corpSigning) SetPendingCLAForLink(linkId, newClaId string) error {
	filter := bson.M{
		fieldLinkId: linkId,
	}

	return impl.dao.UpdateDocsWithoutVersion(filter, bson.M{
		fieldCLANotify:      newClaId,
		fieldClaNotifyCount: 0,
		fieldClaNotifyTime:  0,
	})
}

func (impl *corpSigning) FindPendingAgreements(linkId string) ([]repository.CorpSigningSummary, error) {
	filter := bson.M{
		fieldLinkId:    linkId,
		fieldCLANotify: bson.M{"$ne": ""},
	}

	var dos []corpSigningDO
	err := impl.dao.GetDocs(filter, nil, &dos)
	if err != nil || len(dos) == 0 {
		return nil, err
	}

	r := make([]repository.CorpSigningSummary, len(dos))
	for i := range dos {
		r[i] = dos[i].toCorpSigningSummary()
	}

	return r, nil
}
