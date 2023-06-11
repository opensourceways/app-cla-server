package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

func elemFilterOfCorpSigning(email string) bson.M {
	return filterOfCorpID(email)
}

func (c *client) SignCorpCLA(linkID string, info *dbmodels.CorpSigningCreateOpt) dbmodels.IDBError {
	signing := dCorpSigning{
		ID:          newObjectId(),
		CLALanguage: info.CLALanguage,
		CorpID:      genCorpID(info.AdminEmail),
		CorpName:    info.CorporationName,
		AdminEmail:  info.AdminEmail,
		AdminName:   info.AdminName,
		Date:        info.Date,
		SigningInfo: info.Info,
	}
	signing.Domains = []string{signing.CorpID}
	doc, err := structToMap(signing)
	if err != nil {
		return err
	}

	docFilter := docFilterOfSigning(linkID)
	docFilter["$nor"] = bson.A{
		bson.M{fieldSignings + "." + fieldEmail: info.AdminEmail},
		bson.M{fieldSignings: bson.M{"$elemMatch": bson.M{
			fieldCorpID: genCorpID(info.AdminEmail),
			fieldCorp:   info.CorporationName,
		}}},
	}

	f := func(ctx context.Context) dbmodels.IDBError {
		return c.pushArrayElem(ctx, c.corpSigningCollection, fieldSignings, docFilter, doc)
	}

	return withContext1(f)
}

func toCorpSigningListFilter(opt *dbmodels.CorpSigningListOpt) bson.M {
	c := bson.A{}

	if opt.Lang != "" {
		c = append(c, conditionTofilterArray(bson.M{fieldLang: opt.Lang}))
	}

	if opt.EmailDomain != "" {
		c = append(
			c,
			bson.M{"$isArray": fmt.Sprintf("$$this.%s", fieldDomains)},
			bson.M{"$in": bson.A{
				opt.EmailDomain,
				fmt.Sprintf("$$this.%s", fieldDomains),
			}},
		)
	}

	n := len(c)
	if n == 0 {
		return nil
	}

	if n > 1 {
		return bson.M{"$and": c}
	}

	return c[0].(bson.M)
}

func (this *client) ListCorpSignings(linkID string, opt *dbmodels.CorpSigningListOpt) (
	[]dbmodels.CorporationSigningSummary, dbmodels.IDBError,
) {
	project := projectOfCorpSigning()
	filter := make(map[string]func() bson.M)

	if sf := toCorpSigningListFilter(opt); sf != nil {
		filter[fieldSignings] = func() bson.M { return sf }
	}

	if opt.IncludeAdmin {
		filter[fieldCorpManagers] = func() bson.M {
			return conditionTofilterArray(bson.M{fieldRole: dbmodels.RoleAdmin})
		}

		project[memberNameOfCorpManager(fieldEmail)] = 1
	}

	var v []cCorpSigning
	f := func(ctx context.Context) error {
		return this.getArrayElems(
			ctx, this.corpSigningCollection, docFilterOfSigning(linkID),
			project, filter, &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	n := len(signings)
	if n == 0 {
		return nil, nil
	}

	admins := map[string]bool{}
	for _, item := range v[0].Managers {
		admins[item.Email] = true
	}

	r := make([]dbmodels.CorporationSigningSummary, 0, n)
	for i := 0; i < n; i++ {
		r = append(r, dbmodels.CorporationSigningSummary{
			CorporationSigningBasicInfo: *toDBModelCorporationSigningBasicInfo(&signings[i]),
			AdminAdded:                  admins[signings[i].AdminEmail],
		})
	}

	return r, nil
}

func (this *client) GetCorpSigningBasicInfo(si *dbmodels.SigningIndex) (
	*dbmodels.CorporationSigningBasicInfo, dbmodels.IDBError,
) {
	index := newSigningIndex(si)

	var v []cCorpSigning

	f := func(ctx context.Context) error {
		return this.getArrayElem(
			ctx, this.corpSigningCollection, fieldSignings,
			index.docFilterOfSigning(), index.idFilter(),
			projectOfCorpSigning(), &v,
		)
	}

	if err := withContext(f); err != nil {
		return nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, errNoDBRecord
	}

	signings := v[0].Signings
	if len(signings) == 0 {
		return nil, nil
	}

	return toDBModelCorporationSigningBasicInfo(&(signings[0])), nil
}

func (this *client) GetCorpSigningDetail(si *dbmodels.SigningIndex) (*dbmodels.CLAInfo, *dbmodels.CorpSigningCreateOpt, dbmodels.IDBError) {
	index := newSigningIndex(si)

	pipeline := bson.A{
		bson.M{"$match": index.docFilterOfSigning()},
		bson.M{"$project": bson.M{
			fieldCLAInfos: 1,
			fieldSignings: arrayElemFilter(fieldSignings, index.idFilter()),
		}},
		bson.M{"$unwind": "$" + fieldSignings},
		bson.M{"$project": bson.M{
			fieldSignings: 1,
			fieldCLAInfos: arrayElemFilter(
				fieldCLAInfos,
				bson.M{fieldLang: fmt.Sprintf("$%s.%s", fieldSignings, fieldLang)},
			),
		}},
	}

	var v []struct {
		CLAInfos []DCLAInfo   `bson:"cla_infos"`
		Signings dCorpSigning `bson:"signings"`
	}
	f := func(ctx context.Context) error {
		col := this.collection(this.corpSigningCollection)
		cursor, err := col.Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}

		return cursor.All(ctx, &v)
	}

	if err := withContext(f); err != nil {
		return nil, nil, newSystemError(err)
	}

	if len(v) == 0 {
		return nil, nil, errNoDBRecord
	}

	signing := &(v[0].Signings)
	if signing.CLALanguage == "" {
		return nil, nil, nil
	}

	clas := v[0].CLAInfos
	if len(clas) == 0 {
		return nil, nil, nil
	}

	info := &dbmodels.CorpSigningCreateOpt{
		CorporationSigningBasicInfo: *toDBModelCorporationSigningBasicInfo(signing),
		Info:                        signing.SigningInfo,
	}
	cla := &clas[0]
	return &dbmodels.CLAInfo{
		CLAHash: cla.CLAHash,
		Fields:  toModelOfCLAFields(cla.Fields),
	}, info, nil
}

func toDBModelCorporationSigningBasicInfo(cs *dCorpSigning) *dbmodels.CorporationSigningBasicInfo {
	return &dbmodels.CorporationSigningBasicInfo{
		ID:              cs.ID,
		CLALanguage:     cs.CLALanguage,
		AdminEmail:      cs.AdminEmail,
		AdminName:       cs.AdminName,
		CorporationName: cs.CorpName,
		Date:            cs.Date,
	}
}

func projectOfCorpSigning() bson.M {
	return bson.M{
		memberNameOfSignings(fieldEmail): 1,
		memberNameOfSignings(fieldName):  1,
		memberNameOfSignings(fieldCorp):  1,
		memberNameOfSignings(fieldDate):  1,
		memberNameOfSignings(fieldLang):  1,
	}
}
