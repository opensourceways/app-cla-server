package repositoryimpl

import (
	"errors"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestEmployeeSigningDOCollectedSource(t *testing.T) {
	rep := domain.Representative{
		Name:      dp.CreateName("stu"),
		EmailAddr: dp.CreateEmailAddr("stu@uni.edu.cn"),
	}

	t.Run("source is kept by the round trip", func(t *testing.T) {
		es := domain.NewCollectedEmployeeSigning(
			domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")},
			rep, "2026-01-01", domain.AllSingingInfo{"name": "stu"},
		)

		do := toEmployeeSigningDO(&es)

		if do.Source != domain.EmployeeSigningSourceIndividual {
			t.Errorf("DO source = %q, want %q", do.Source, domain.EmployeeSigningSourceIndividual)
		}

		back := do.toEmployeeSigning()

		if back.Source != domain.EmployeeSigningSourceIndividual {
			t.Errorf("domain source = %q, want %q", back.Source, domain.EmployeeSigningSourceIndividual)
		}
		if back.Enabled {
			t.Error("collected employee signing should stay disabled")
		}
		if len(back.Logs) != 1 || back.Logs[0].Action != "collect" {
			t.Errorf("logs = %v, want one collect entry", back.Logs)
		}
	})

	t.Run("source is omitted when empty", func(t *testing.T) {
		es := domain.EmployeeSigning{
			Id:   "e1",
			CLA:  domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")},
			Rep:  rep,
			Date: "2026-01-01",
		}

		do := toEmployeeSigningDO(&es)
		doc, err := do.toDoc()
		if err != nil {
			t.Fatalf("toDoc: unexpected error %v", err)
		}

		if _, ok := doc["source"]; ok {
			t.Errorf("source should be omitted for self-signed employee signing, got %v", doc["source"])
		}
	})
}

func TestToDomainRegex(t *testing.T) {
	got := toDomainRegex([]string{"uni.edu.cn", "a+b.cn", "x.*y.org"})

	regex, ok := got["$regex"].(string)
	if !ok {
		t.Fatalf("regex = %#v, want a string", got["$regex"])
	}

	expect := `(?i)^(uni\.edu\.cn|a\+b\.cn|x\.\*y\.org)$`
	if regex != expect {
		t.Errorf("regex = %q, want %q", regex, expect)
	}
}

func TestIndividualSigningIndexes(t *testing.T) {
	indexes := IndividualSigningIndexes()

	if len(indexes) != 1 {
		t.Fatalf("indexes = %d, want 1", len(indexes))
	}

	keys, ok := indexes[0].Keys.(bson.D)
	if !ok {
		t.Fatalf("index keys type = %T, want bson.D", indexes[0].Keys)
	}
	if len(keys) != 2 {
		t.Fatalf("index keys = %d, want 2", len(keys))
	}
	if keys[0].Key != fieldLinkId || keys[1].Key != fieldDomain {
		t.Errorf("index keys = %v, want link_id+domain", keys)
	}
}

// fakeDAO implements dao by embedding the interface, only the methods used
// by the tests are overridden.
type fakeDAO struct {
	dao

	getDocsFn                  func(filter, project bson.M, result interface{}) error
	updateDocsWithoutVersionFn func(filter, doc bson.M) error
}

func (f *fakeDAO) GetDocs(filter, project bson.M, result interface{}) error {
	return f.getDocsFn(filter, project, result)
}

func (f *fakeDAO) UpdateDocsWithoutVersion(filter, doc bson.M) error {
	return f.updateDocsWithoutVersionFn(filter, doc)
}

func TestIndividualSigningFindByDomains(t *testing.T) {
	t.Run("no domains", func(t *testing.T) {
		impl := NewIndividualSigning(&fakeDAO{})

		v, err := impl.FindByDomains("link1", nil)
		if err != nil || v != nil {
			t.Errorf("got (%v, %v), want (nil, nil)", v, err)
		}
	})

	t.Run("builds the expected filter and converts the docs", func(t *testing.T) {
		var gotFilter bson.M

		impl := NewIndividualSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				gotFilter = filter

				dos, ok := result.(*[]individualSigningDO)
				if !ok {
					t.Fatalf("result type = %T, want *[]individualSigningDO", result)
				}

				*dos = []individualSigningDO{{
					LinkId: "link1",
					Date:   "2026-01-01",
					Domain: "uni.edu.cn",
				}}

				return nil
			},
		})

		v, err := impl.FindByDomains("link1", []string{"UNI.edu.cn"})
		if err != nil {
			t.Fatalf("FindByDomains: unexpected error %v", err)
		}

		if gotFilter[fieldLinkId] != "link1" {
			t.Errorf("filter link_id = %v, want link1", gotFilter[fieldLinkId])
		}
		if gotFilter[fieldDeleted] != false {
			t.Errorf("filter deleted = %v, want false", gotFilter[fieldDeleted])
		}

		domainCond, ok := gotFilter[fieldDomain].(bson.M)
		if !ok {
			t.Fatalf("filter domain type = %T, want bson.M", gotFilter[fieldDomain])
		}
		regex, ok := domainCond["$regex"].(string)
		if !ok || regex != `(?i)^(UNI\.edu\.cn)$` {
			t.Errorf("domain regex = %#v, want an anchored case-insensitive regex", domainCond["$regex"])
		}

		if len(v) != 1 || v[0].Link.Id != "link1" || v[0].Date != "2026-01-01" {
			t.Errorf("result = %+v, want the converted individual signing", v)
		}
	})

	t.Run("propagates the dao error", func(t *testing.T) {
		impl := NewIndividualSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				return errors.New("db error")
			},
		})

		if _, err := impl.FindByDomains("link1", []string{"uni.edu.cn"}); err == nil {
			t.Error("FindByDomains should propagate the dao error")
		}
	})
}

func TestIndividualSigningRemoveAll(t *testing.T) {
	t.Run("soft deletes with the same case-insensitive domain conditions", func(t *testing.T) {
		var gotFilter, gotDoc bson.M

		impl := NewIndividualSigning(&fakeDAO{
			updateDocsWithoutVersionFn: func(filter, doc bson.M) error {
				gotFilter = filter
				gotDoc = doc
				return nil
			},
		})

		if err := impl.RemoveAll("link1", []string{"UNI.edu.cn"}); err != nil {
			t.Fatalf("RemoveAll: unexpected error %v", err)
		}

		if gotFilter[fieldLinkId] != "link1" {
			t.Errorf("filter link_id = %v, want link1", gotFilter[fieldLinkId])
		}
		if gotFilter[fieldDeleted] != false {
			t.Errorf("filter deleted = %v, want false", gotFilter[fieldDeleted])
		}

		domainCond, ok := gotFilter[fieldDomain].(bson.M)
		if !ok {
			t.Fatalf("filter domain type = %T, want bson.M", gotFilter[fieldDomain])
		}
		conditions, ok := domainCond[mongodbCmdIn].(bson.A)
		if !ok || len(conditions) != 1 {
			t.Fatalf("domain conditions = %#v, want one", domainCond[mongodbCmdIn])
		}
		regex, ok := conditions[0].(bson.M)
		if !ok || regex["$regex"] != "(?i)^UNI\\.edu\\.cn$" {
			t.Errorf("first condition = %#v, want an anchored case-insensitive regex, not a case-sensitive $in", conditions[0])
		}

		if gotDoc[fieldDeleted] != true {
			t.Errorf("update deleted = %v, want true", gotDoc[fieldDeleted])
		}
		if _, ok := gotDoc[fieldDeletedAt]; !ok {
			t.Errorf("update doc = %v, want deleted_at to be set", gotDoc)
		}
	})

	t.Run("propagates the dao error", func(t *testing.T) {
		impl := NewIndividualSigning(&fakeDAO{
			updateDocsWithoutVersionFn: func(filter, doc bson.M) error {
				return errors.New("db error")
			},
		})

		if err := impl.RemoveAll("link1", []string{"uni.edu.cn"}); err == nil {
			t.Error("RemoveAll should propagate the dao error")
		}
	})
}

// TestRemoveAllAndFindByDomainsShareDomainMatching guards the invariant that
// the soft delete of RemoveAll matches exactly the records that
// FindByDomains collects: with mixed-case domains (corp registers
// "UNI.edu.cn" while individual signings store "uni.edu.cn"), a record must
// never be collected without being soft deleted afterwards.
func TestRemoveAllAndFindByDomainsShareDomainMatching(t *testing.T) {
	domains := []string{"UNI.edu.cn", "Corp.IO", "x.y.org"}

	var findFilter, removeFilter bson.M

	impl := NewIndividualSigning(&fakeDAO{
		getDocsFn: func(filter, project bson.M, result interface{}) error {
			findFilter = filter
			return nil
		},
		updateDocsWithoutVersionFn: func(filter, doc bson.M) error {
			removeFilter = filter
			return nil
		},
	})

	if _, err := impl.FindByDomains("link1", domains); err != nil {
		t.Fatalf("FindByDomains: unexpected error %v", err)
	}
	if err := impl.RemoveAll("link1", domains); err != nil {
		t.Fatalf("RemoveAll: unexpected error %v", err)
	}

	if !reflect.DeepEqual(findFilter[fieldDomain], removeFilter[fieldDomain]) {
		t.Errorf("domain conditions differ: FindByDomains = %#v, RemoveAll = %#v",
			findFilter[fieldDomain], removeFilter[fieldDomain])
	}
}
