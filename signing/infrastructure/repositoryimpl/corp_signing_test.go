package repositoryimpl

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

// --- buildCorpSigningSearchFilter ---

func TestBuildCorpSigningSearchFilter(t *testing.T) {
	t.Run("escapes regex metacharacters", func(t *testing.T) {
		// regexp.QuoteMeta escapes ASCII regex metacharacters only.
		// Full-width characters like （） are not special in RE2 and are
		// returned as-is, which is correct behaviour.
		cases := map[string]string{
			`A.B*C`:    `A\.B\*C`,
			`test(x)`:  `test\(x\)`,
			`a+b`:      `a\+b`,
			`[bracket]`: `\[bracket\]`,
			`{brace}`:  `\{brace\}`,
			`dot.dot`:  `dot\.dot`,
		}
		for q, wantRegex := range cases {
			got := buildCorpSigningSearchFilter(q)
			orVal, ok := got[mongodbCmdOr].(bson.A)
			if !ok || len(orVal) != 2 {
				t.Fatalf("q=%q: expected $or array of 2, got %#v", q, got[mongodbCmdOr])
			}

			for _, branch := range orVal {
				branchM, ok := branch.(bson.M)
				if !ok {
					t.Fatalf("q=%q: branch is not bson.M", q)
				}
				if len(branchM) != 1 {
					t.Fatalf("q=%q: branch should have exactly 1 key, got %d", q, len(branchM))
				}
				for _, cond := range branchM {
					condM, ok := cond.(bson.M)
					if !ok {
						t.Fatalf("q=%q: condition is not bson.M", q)
					}
					regex, ok := condM[mongodbCmdRegex].(string)
					if !ok {
						t.Fatalf("q=%q: $regex is not a string", q)
					}
					if regex != wantRegex {
						t.Errorf("q=%q: regex = %q, want %q", q, regex, wantRegex)
					}
				}
			}
		}
	})

	t.Run("contains corp.name and rep.email branches", func(t *testing.T) {
		got := buildCorpSigningSearchFilter("华为")
		orVal := got[mongodbCmdOr].(bson.A)
		keys := make([]string, 0, 2)
		for _, branch := range orVal {
			for k := range branch.(bson.M) {
				keys = append(keys, k)
			}
		}
		want := []string{childField(fieldCorp, fieldName), childField(fieldRep, fieldEmail)}
		if !reflect.DeepEqual(keys, want) {
			t.Errorf("branch keys = %v, want %v", keys, want)
		}
	})

	t.Run("sets case-insensitive option", func(t *testing.T) {
		got := buildCorpSigningSearchFilter("test")
		orVal := got[mongodbCmdOr].(bson.A)
		for _, branch := range orVal {
			cond := branch.(bson.M)
			for _, c := range cond {
				opts := c.(bson.M)["$options"]
				if opts != "i" {
					t.Errorf("$options = %v, want i", opts)
				}
			}
		}
	})
}

// --- buildCorpSigningPagePipeline ---

func TestBuildCorpSigningPagePipeline(t *testing.T) {
	filter := bson.M{"link_id": "l1"}
	project := corpSigningPageProject()
	pipeline := buildCorpSigningPagePipeline(filter, project, 2, 10)

	t.Run("starts with $match", func(t *testing.T) {
		first, ok := pipeline[0].(bson.M)
		if !ok {
			t.Fatalf("pipeline[0] is not bson.M")
		}
		if _, ok := first["$match"]; !ok {
			t.Errorf("pipeline[0] = %v, want $match", first)
		}
	})

	t.Run("data branch has $sort before $skip", func(t *testing.T) {
		facet := pipeline[1].(bson.M)["$facet"].(bson.M)
		data := facet["data"].(bson.A)

		var stageNames []string
		for _, stage := range data {
			for k := range stage.(bson.M) {
				stageNames = append(stageNames, k)
			}
		}

		want := []string{"$sort", "$skip", "$limit", "$project"}
		if !reflect.DeepEqual(stageNames, want) {
			t.Errorf("data stages = %v, want %v", stageNames, want)
		}
	})

	t.Run("$sort uses date desc and _id desc", func(t *testing.T) {
		facet := pipeline[1].(bson.M)["$facet"].(bson.M)
		data := facet["data"].(bson.A)
		sortStage := data[0].(bson.M)["$sort"].(bson.D)
		if len(sortStage) != 2 {
			t.Fatalf("sort keys = %d, want 2", len(sortStage))
		}
		if sortStage[0].Key != fieldDate || sortStage[0].Value != -1 {
			t.Errorf("sort[0] = %v, want date:-1", sortStage[0])
		}
		if sortStage[1].Key != "_id" || sortStage[1].Value != -1 {
			t.Errorf("sort[1] = %v, want _id:-1", sortStage[1])
		}
	})

	t.Run("total branch has $count", func(t *testing.T) {
		facet := pipeline[1].(bson.M)["$facet"].(bson.M)
		total := facet["total"].(bson.A)
		if len(total) != 1 {
			t.Fatalf("total stages = %d, want 1", len(total))
		}
		if _, ok := total[0].(bson.M)["$count"]; !ok {
			t.Errorf("total[0] = %v, want $count", total[0])
		}
	})
}

// --- corpSigningPageProject ---

func TestCorpSigningPageProject(t *testing.T) {
	p := corpSigningPageProject()
	if _, ok := p[fieldAdminAddedDate]; !ok {
		t.Errorf("project missing %s", fieldAdminAddedDate)
	}
}

// --- CorpSigningIndexes ---

func TestCorpSigningIndexes(t *testing.T) {
	indexes := CorpSigningIndexes()

	found := false
	for _, idx := range indexes {
		keys, ok := idx.Keys.(bson.D)
		if !ok {
			continue
		}
		if len(keys) == 3 &&
			keys[0].Key == fieldLinkId && keys[0].Value == 1 &&
			keys[1].Key == fieldDate && keys[1].Value == -1 &&
			keys[2].Key == "_id" && keys[2].Value == -1 {
			found = true
		}
	}
	if !found {
		t.Errorf("compound index {link_id:1, date:-1, _id:-1} not found in %d indexes", len(indexes))
	}
}

// --- corpSigningDO.toCorpSigningSummary ---

func TestCorpSigningDOToCorpSigningSummaryAdminAddedDate(t *testing.T) {
	t.Run("maps admin_added_date when present", func(t *testing.T) {
		do := corpSigningDO{
			Id:             primitive.NewObjectID(),
			Date:           "2026-08-30",
			AdminAddedDate: "2026-08-31",
		}
		s := do.toCorpSigningSummary()
		if s.AdminAddedDate != "2026-08-31" {
			t.Errorf("AdminAddedDate = %q, want 2026-08-31", s.AdminAddedDate)
		}
	})

	t.Run("admin_added_date empty when absent", func(t *testing.T) {
		do := corpSigningDO{
			Id:   primitive.NewObjectID(),
			Date: "2026-08-30",
		}
		s := do.toCorpSigningSummary()
		if s.AdminAddedDate != "" {
			t.Errorf("AdminAddedDate = %q, want empty", s.AdminAddedDate)
		}
	})
}

// --- formatObjectIDDate ---

func TestFormatObjectIDDate(t *testing.T) {
	t.Run("valid object id timestamp", func(t *testing.T) {
		// Create an ObjectID from a known timestamp
		ts := time.Unix(1725062400, 0)
		id := primitive.NewObjectIDFromTimestamp(ts)
		// The exact date depends on timezone; just check it's non-empty and formatted
		got := formatObjectIDDate(id)
		if got == "" {
			t.Error("expected non-empty date")
		}
		if len(got) != 10 {
			t.Errorf("date length = %d, want 10 (YYYY-MM-DD)", len(got))
		}
	})

	t.Run("zero object id returns epoch date", func(t *testing.T) {
		got := formatObjectIDDate(primitive.NilObjectID)
		if got == "" {
			t.Error("expected non-empty date for zero ObjectID")
		}
		if len(got) != 10 {
			t.Errorf("date length = %d, want 10", len(got))
		}
	})
}

// --- deriveAdminAddedDate ---

func TestDeriveAdminAddedDate(t *testing.T) {
	csId := primitive.NewObjectID()

	t.Run("uses user _id timestamp when user found", func(t *testing.T) {
		doc := &corpSigningDO{
			Id:    csId,
			Admin: managerDO{Id: "admin1"},
		}

		userTs := time.Unix(1725148800, 0)
		userObjId := primitive.NewObjectIDFromTimestamp(userTs)
		wantUserDate := formatObjectIDDate(userObjId)

		d := &backfillFakeDAO{
			getDocFn: func(filter, project bson.M, result interface{}) error {
				u, ok := result.(*userDO)
				if !ok {
					t.Fatalf("result type = %T, want *userDO", result)
				}
				u.Id = userObjId
				return nil
			},
		}

		got := deriveAdminAddedDate(doc, d)
		if got != wantUserDate {
			t.Errorf("deriveAdminAddedDate = %q, want %q (user date)", got, wantUserDate)
		}
	})

	t.Run("falls back to corp_signing _id when user not found", func(t *testing.T) {
		doc := &corpSigningDO{
			Id:    csId,
			Admin: managerDO{Id: "admin1"},
		}
		wantDate := formatObjectIDDate(csId)

		d := &backfillFakeDAO{
			getDocFn: func(filter, project bson.M, result interface{}) error {
				return errors.New("not found")
			},
		}

		got := deriveAdminAddedDate(doc, d)
		if got != wantDate {
			t.Errorf("deriveAdminAddedDate = %q, want %q (corp_signing fallback)", got, wantDate)
		}
	})
}

// --- BackfillAdminAddedDate ---

func TestBackfillAdminAddedDate(t *testing.T) {
	t.Run("skips when no docs need backfill", func(t *testing.T) {
		corpDao := &backfillFakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos, _ := result.(*[]corpSigningDO)
				*dos = []corpSigningDO{}
				return nil
			},
		}

		if err := BackfillAdminAddedDate(corpDao, &backfillFakeDAO{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("updates docs missing admin_added_date", func(t *testing.T) {
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		var updatedFilters []bson.M
		var updatedDocs []bson.M

		corpDao := &backfillFakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos, _ := result.(*[]corpSigningDO)
				*dos = []corpSigningDO{
					{Id: id1, Admin: managerDO{Id: "admin1"}},
					{Id: id2, Admin: managerDO{Id: "admin2"}},
				}
				return nil
			},
			updateDocsWithoutVersionFn: func(filter, doc bson.M) error {
				updatedFilters = append(updatedFilters, filter)
				updatedDocs = append(updatedDocs, doc)
				return nil
			},
		}
		userDao := &backfillFakeDAO{
			getDocFn: func(filter, project bson.M, result interface{}) error {
				return errors.New("not found")
			},
		}

		if err := BackfillAdminAddedDate(corpDao, userDao); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(updatedFilters) != 2 {
			t.Fatalf("updated %d docs, want 2", len(updatedFilters))
		}

		for i, doc := range updatedDocs {
			date, ok := doc[fieldAdminAddedDate].(string)
			if !ok || date == "" {
				t.Errorf("update %d: admin_added_date = %v, want non-empty string", i, doc[fieldAdminAddedDate])
			}
			if len(date) != 10 {
				t.Errorf("update %d: date = %q, want YYYY-MM-DD format", i, date)
			}
		}
	})

	t.Run("query error does not return error", func(t *testing.T) {
		corpDao := &backfillFakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				return errors.New("db connection lost")
			},
		}

		if err := BackfillAdminAddedDate(corpDao, &backfillFakeDAO{}); err != nil {
			t.Errorf("expected nil error on query failure, got %v", err)
		}
	})

	t.Run("update error continues processing remaining docs", func(t *testing.T) {
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()
		callCount := 0

		corpDao := &backfillFakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos, _ := result.(*[]corpSigningDO)
				*dos = []corpSigningDO{
					{Id: id1, Admin: managerDO{Id: "admin1"}},
					{Id: id2, Admin: managerDO{Id: "admin2"}},
				}
				return nil
			},
			updateDocsWithoutVersionFn: func(filter, doc bson.M) error {
				callCount++
				if callCount == 1 {
					return errors.New("write timeout")
				}
				return nil
			},
		}
		userDao := &backfillFakeDAO{
			getDocFn: func(filter, project bson.M, result interface{}) error {
				return errors.New("not found")
			},
		}

		_ = BackfillAdminAddedDate(corpDao, userDao)
		if callCount != 2 {
			t.Errorf("update called %d times, want 2 (should continue after first failure)", callCount)
		}
	})
}

// --- FindPage ---

func TestFindPageFilterConstruction(t *testing.T) {
	t.Run("adminAdded true with search uses direct field and $or search", func(t *testing.T) {
		var gotFilter bson.M
		var gotPipeline bson.A

		impl := &corpSigning{dao: &findPageFakeDAO{
			aggregateFn: func(pipeline bson.A, result interface{}) error {
				gotPipeline = pipeline
				return nil
			},
		}}

		_, _ = impl.FindPage("link1", 1, 10, true, "华为")

		// Extract the filter from the $match stage
		matchStage := gotPipeline[0].(bson.M)
		gotFilter = matchStage["$match"].(bson.M)

		if gotFilter[fieldLinkId] != "link1" {
			t.Errorf("filter link_id = %v, want link1", gotFilter[fieldLinkId])
		}
		if !reflect.DeepEqual(gotFilter["admin.id"], bson.M{"$ne": ""}) {
			t.Errorf("filter admin.id = %v, want $ne empty", gotFilter["admin.id"])
		}
		orVal, ok := gotFilter[mongodbCmdOr].(bson.A)
		if !ok || len(orVal) != 2 {
			t.Errorf("filter $or = %v, want array of 2 branches", gotFilter[mongodbCmdOr])
		}
		// Should NOT have $and when only search uses $or
		if _, ok := gotFilter[mongodbCmdAnd]; ok {
			t.Errorf("filter should not have $and when adminAdded=true")
		}
	})

	t.Run("adminAdded false with search wraps both $or in $and", func(t *testing.T) {
		var gotPipeline bson.A

		impl := &corpSigning{dao: &findPageFakeDAO{
			aggregateFn: func(pipeline bson.A, result interface{}) error {
				gotPipeline = pipeline
				return nil
			},
		}}

		_, _ = impl.FindPage("link1", 1, 10, false, "华为")

		matchStage := gotPipeline[0].(bson.M)
		gotFilter := matchStage["$match"].(bson.M)

		andVal, ok := gotFilter[mongodbCmdAnd].(bson.A)
		if !ok {
			t.Fatalf("filter should have $and when both admin and search use $or")
		}
		if len(andVal) != 2 {
			t.Fatalf("$and has %d conditions, want 2", len(andVal))
		}

		// First condition: admin $or (admin.id == "" or admin missing)
		adminOr := andVal[0].(bson.M)
		if _, ok := adminOr[mongodbCmdOr]; !ok {
			t.Errorf("first $and condition should be admin $or")
		}

		// Second condition: search $or
		searchOr := andVal[1].(bson.M)
		if _, ok := searchOr[mongodbCmdOr]; !ok {
			t.Errorf("second $and condition should be search $or")
		}
	})

	t.Run("adminAdded false without search uses $or directly", func(t *testing.T) {
		var gotPipeline bson.A

		impl := &corpSigning{dao: &findPageFakeDAO{
			aggregateFn: func(pipeline bson.A, result interface{}) error {
				gotPipeline = pipeline
				return nil
			},
		}}

		_, _ = impl.FindPage("link1", 1, 10, false, "")

		matchStage := gotPipeline[0].(bson.M)
		gotFilter := matchStage["$match"].(bson.M)

		if _, ok := gotFilter[mongodbCmdOr]; !ok {
			t.Errorf("filter should have $or for admin condition")
		}
		if _, ok := gotFilter[mongodbCmdAnd]; ok {
			t.Errorf("filter should not have $and when no search")
		}
	})

	t.Run("adminAdded true without search uses direct field", func(t *testing.T) {
		var gotPipeline bson.A

		impl := &corpSigning{dao: &findPageFakeDAO{
			aggregateFn: func(pipeline bson.A, result interface{}) error {
				gotPipeline = pipeline
				return nil
			},
		}}

		_, _ = impl.FindPage("link1", 1, 10, true, "")

		matchStage := gotPipeline[0].(bson.M)
		gotFilter := matchStage["$match"].(bson.M)

		if !reflect.DeepEqual(gotFilter["admin.id"], bson.M{"$ne": ""}) {
			t.Errorf("filter admin.id = %v, want $ne empty", gotFilter["admin.id"])
		}
		if _, ok := gotFilter[mongodbCmdOr]; ok {
			t.Errorf("filter should not have $or when no search")
		}
		if _, ok := gotFilter[mongodbCmdAnd]; ok {
			t.Errorf("filter should not have $and when no search")
		}
	})

	t.Run("empty search omits search filter", func(t *testing.T) {
		var gotPipeline bson.A

		impl := &corpSigning{dao: &findPageFakeDAO{
			aggregateFn: func(pipeline bson.A, result interface{}) error {
				gotPipeline = pipeline
				return nil
			},
		}}

		_, _ = impl.FindPage("link1", 1, 10, true, "")

		matchStage := gotPipeline[0].(bson.M)
		gotFilter := matchStage["$match"].(bson.M)

		// Should only have link_id and admin.id
		if len(gotFilter) != 2 {
			t.Errorf("filter has %d keys, want 2 (link_id + admin.id)", len(gotFilter))
		}
	})

	t.Run("maps result data to summaries with admin_added_date", func(t *testing.T) {
		impl := &corpSigning{dao: &findPageFakeDAO{
			aggregateFn: func(pipeline bson.A, result interface{}) error {
				facet, ok := result.(*[]struct {
					Total []struct {
						N int64 `bson:"n"`
					} `bson:"total"`
					Data []corpSigningDO `bson:"data"`
				})
				if !ok {
					t.Fatalf("result type = %T", result)
				}
				*facet = []struct {
					Total []struct {
						N int64 `bson:"n"`
					} `bson:"total"`
					Data []corpSigningDO `bson:"data"`
				}{{
					Total: []struct {
						N int64 `bson:"n"`
					}{{N: 1}},
					Data: []corpSigningDO{{
						Id:             primitive.NewObjectID(),
						Date:           "2026-08-30",
						AdminAddedDate: "2026-08-31",
					}},
				}}
				return nil
			},
		}}

		page, err := impl.FindPage("link1", 1, 10, true, "")
		if err != nil {
			t.Fatalf("FindPage error: %v", err)
		}
		if page.Total != 1 {
			t.Fatalf("Total = %d, want 1", page.Total)
		}
		if len(page.Data) != 1 {
			t.Fatalf("Data len = %d, want 1", len(page.Data))
		}
		if page.Data[0].AdminAddedDate != "2026-08-31" {
			t.Errorf("AdminAddedDate = %q, want 2026-08-31", page.Data[0].AdminAddedDate)
		}
	})
}

// --- AddAdmin ---

func TestAddAdminWritesAdminAddedDate(t *testing.T) {
	var gotDoc bson.M

	csId := primitive.NewObjectID()
	hex := csId.Hex()

	impl := &corpSigning{dao: &addAdminFakeDAO{
		docIdFilterFn: func(s string) (bson.M, error) {
			oid, _ := primitive.ObjectIDFromHex(hex)
			return bson.M{"_id": oid}, nil
		},
		updateDocFn: func(filter, doc bson.M, version int) error {
			gotDoc = doc
			return nil
		},
	}}

	admin := domain.Manager{
		Id: "admin1",
		Representative: domain.Representative{
			Name:      dp.CreateName("张三"),
			EmailAddr: dp.CreateEmailAddr("zhangsan@test.com"),
		},
	}

	cs := &domain.CorpSigning{
		Id:     hex,
		Admin:  admin,
		Link:   domain.LinkInfo{Id: "link1"},
	}

	if err := impl.AddAdmin(cs); err != nil {
		t.Fatalf("AddAdmin error: %v", err)
	}

	if _, ok := gotDoc[fieldAdmin]; !ok {
		t.Errorf("update doc missing %s", fieldAdmin)
	}
	date, ok := gotDoc[fieldAdminAddedDate].(string)
	if !ok || date == "" {
		t.Errorf("update doc %s = %v, want non-empty date string", fieldAdminAddedDate, gotDoc[fieldAdminAddedDate])
	}
	if len(date) != 10 {
		t.Errorf("date = %q, want YYYY-MM-DD format (10 chars)", date)
	}
}

func TestAddAdminConcurrentUpdate(t *testing.T) {
	concurrentErr := errors.New("doc doesn't exist")

	impl := &corpSigning{dao: &addAdminFakeDAO{
		docIdFilterFn: func(s string) (bson.M, error) {
			return bson.M{"_id": primitive.NewObjectID()}, nil
		},
		updateDocFn: func(filter, doc bson.M, version int) error {
			return concurrentErr
		},
		isDocNotExistsFn: func(err error) bool {
			return err == concurrentErr
		},
	}}

	cs := &domain.CorpSigning{
		Id:    "507f1f77bcf86cd799439011",
		Admin: domain.Manager{Id: "admin1"},
		Link:  domain.LinkInfo{Id: "link1"},
	}

	err := impl.AddAdmin(cs)
	if err == nil {
		t.Error("expected error on concurrent update")
	}
}

func TestAddAdminInvalidId(t *testing.T) {
	impl := &corpSigning{dao: &addAdminFakeDAO{
		docIdFilterFn: func(s string) (bson.M, error) {
			oid, err := primitive.ObjectIDFromHex(s)
			if err != nil {
				return nil, err
			}
			return bson.M{"_id": oid}, nil
		},
	}}

	cs := &domain.CorpSigning{
		Id:    "invalid-hex",
		Admin: domain.Manager{Id: "admin1"},
		Link:  domain.LinkInfo{Id: "link1"},
	}

	err := impl.AddAdmin(cs)
	if err == nil {
		t.Error("expected error for invalid ObjectID hex")
	}
}

// --- fake DAOs for tests ---

// findPageFakeDAO supports Aggregate only.
type findPageFakeDAO struct {
	dao
	aggregateFn func(pipeline bson.A, result interface{}) error
}

func (f *findPageFakeDAO) Aggregate(pipeline bson.A, result interface{}) error {
	return f.aggregateFn(pipeline, result)
}

// addAdminFakeDAO supports DocIdFilter and UpdateDoc.
type addAdminFakeDAO struct {
	dao
	docIdFilterFn    func(s string) (bson.M, error)
	updateDocFn      func(filter, doc bson.M, version int) error
	isDocNotExistsFn func(err error) bool
}

func (f *addAdminFakeDAO) DocIdFilter(s string) (bson.M, error) {
	return f.docIdFilterFn(s)
}

func (f *addAdminFakeDAO) UpdateDoc(filter, doc bson.M, version int) error {
	return f.updateDocFn(filter, doc, version)
}

func (f *addAdminFakeDAO) IsDocNotExists(err error) bool {
	if f.isDocNotExistsFn != nil {
		return f.isDocNotExistsFn(err)
	}
	return false
}

// Ensure unused mongo import is referenced for index test type assertion.
var _ = mongo.IndexModel{}

// backfillFakeDAO supports GetDocs, GetDoc, and UpdateDocsWithoutVersion.
type backfillFakeDAO struct {
	dao
	getDocsFn                  func(filter, project bson.M, result interface{}) error
	getDocFn                   func(filter, project bson.M, result interface{}) error
	updateDocsWithoutVersionFn func(filter, doc bson.M) error
}

func (f *backfillFakeDAO) GetDocs(filter, project bson.M, result interface{}) error {
	return f.getDocsFn(filter, project, result)
}

func (f *backfillFakeDAO) GetDoc(filter, project bson.M, result interface{}) error {
	return f.getDocFn(filter, project, result)
}

func (f *backfillFakeDAO) UpdateDocsWithoutVersion(filter, doc bson.M) error {
	return f.updateDocsWithoutVersionFn(filter, doc)
}
