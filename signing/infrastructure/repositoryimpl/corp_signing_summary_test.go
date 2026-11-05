package repositoryimpl

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCorpSigningFindCorpSummary covers the infrastructure-layer FindCorpSummary,
// whose projection (fieldAdmin) and HasAdmin flag (item.Admin.Email != "") were
// changed in this round. It reuses the in-package fakeDAO so no real MongoDB is
// required.
func TestCorpSigningFindCorpSummary(t *testing.T) {
	t.Run("builds the expected filter and projection", func(t *testing.T) {
		var gotFilter, gotProject bson.M

		impl := NewCorpSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				gotFilter = filter
				gotProject = project

				dos, ok := result.(*[]corpSigningDO)
				if !ok {
					t.Fatalf("result type = %T, want *[]corpSigningDO", result)
				}
				*dos = nil
				return nil
			},
		})

		if _, err := impl.FindCorpSummary("link1", "example.com"); err != nil {
			t.Fatalf("FindCorpSummary: unexpected error %v", err)
		}

		if gotFilter[fieldLinkId] != "link1" {
			t.Errorf("filter link_id = %v, want link1", gotFilter[fieldLinkId])
		}

		domainField := childField(fieldCorp, fieldDomains)
		domainCond, ok := gotFilter[domainField].(bson.M)
		if !ok {
			t.Fatalf("filter %s type = %T, want bson.M", domainField, gotFilter[domainField])
		}
		inArr, ok := domainCond[mongodbCmdIn].(bson.A)
		if !ok || len(inArr) != 1 || inArr[0] != "example.com" {
			t.Errorf("filter %s $in = %#v, want [example.com]", domainField, domainCond[mongodbCmdIn])
		}

		// The projection must project corp and admin (the super admin), not the
		// legacy managers array.
		if gotProject[fieldCorp] != 1 || gotProject[fieldAdmin] != 1 {
			t.Errorf("project = %#v, want {corp:1, admin:1}", gotProject)
		}
		if _, ok := gotProject[fieldManagers]; ok {
			t.Errorf("project must not contain managers, got %#v", gotProject)
		}
	})

	t.Run("HasAdmin is true when admin email is set", func(t *testing.T) {
		id := primitive.NewObjectID()

		impl := NewCorpSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos := result.(*[]corpSigningDO)
				*dos = []corpSigningDO{{
					Id:    id,
					Corp:  corpDO{Name: "approved corp", Domain: "example.com", Domains: []string{"example.com"}},
					Admin: managerDO{Id: "admin1", RepDO: RepDO{Email: "admin@example.com"}},
				}}
				return nil
			},
		})

		v, err := impl.FindCorpSummary("link1", "example.com")
		if err != nil {
			t.Fatalf("FindCorpSummary: unexpected error %v", err)
		}
		if len(v) != 1 {
			t.Fatalf("result count: got %d, want 1", len(v))
		}
		if !v[0].HasAdmin {
			t.Error("HasAdmin should be true when admin email is set")
		}
		if v[0].CorpName.CorpName() != "approved corp" {
			t.Errorf("CorpName = %q, want approved corp", v[0].CorpName.CorpName())
		}
		if v[0].CorpSigningId != id.Hex() {
			t.Errorf("CorpSigningId = %q, want %q", v[0].CorpSigningId, id.Hex())
		}
	})

	t.Run("HasAdmin is false when admin email is empty", func(t *testing.T) {
		id := primitive.NewObjectID()

		impl := NewCorpSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos := result.(*[]corpSigningDO)
				*dos = []corpSigningDO{{
					Id:    id,
					Corp:  corpDO{Name: "pending corp", Domain: "example.com", Domains: []string{"example.com"}},
					Admin: managerDO{Id: "", RepDO: RepDO{Email: ""}},
				}}
				return nil
			},
		})

		v, err := impl.FindCorpSummary("link1", "example.com")
		if err != nil {
			t.Fatalf("FindCorpSummary: unexpected error %v", err)
		}
		if len(v) != 1 {
			t.Fatalf("result count: got %d, want 1", len(v))
		}
		if v[0].HasAdmin {
			t.Error("HasAdmin should be false when admin email is empty")
		}
		if v[0].CorpSigningId != id.Hex() {
			t.Errorf("CorpSigningId = %q, want %q", v[0].CorpSigningId, id.Hex())
		}
	})

	t.Run("returns multiple summaries preserving order", func(t *testing.T) {
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		impl := NewCorpSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos := result.(*[]corpSigningDO)
				*dos = []corpSigningDO{
					{Id: id1, Corp: corpDO{Name: "corp a"}, Admin: managerDO{RepDO: RepDO{Email: "a@x.com"}}},
					{Id: id2, Corp: corpDO{Name: "corp b"}, Admin: managerDO{}},
				}
				return nil
			},
		})

		v, err := impl.FindCorpSummary("link1", "x.com")
		if err != nil {
			t.Fatalf("FindCorpSummary: unexpected error %v", err)
		}
		if len(v) != 2 {
			t.Fatalf("result count: got %d, want 2", len(v))
		}
		if v[0].CorpName.CorpName() != "corp a" || !v[0].HasAdmin || v[0].CorpSigningId != id1.Hex() {
			t.Errorf("v[0] = %+v, want corp a/true/%s", v[0], id1.Hex())
		}
		if v[1].CorpName.CorpName() != "corp b" || v[1].HasAdmin || v[1].CorpSigningId != id2.Hex() {
			t.Errorf("v[1] = %+v, want corp b/false/%s", v[1], id2.Hex())
		}
	})

	t.Run("returns empty slice when no docs", func(t *testing.T) {
		impl := NewCorpSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				dos := result.(*[]corpSigningDO)
				*dos = nil
				return nil
			},
		})

		v, err := impl.FindCorpSummary("link1", "example.com")
		if err != nil {
			t.Fatalf("FindCorpSummary: unexpected error %v", err)
		}
		if len(v) != 0 {
			t.Errorf("result count: got %d, want 0", len(v))
		}
	})

	t.Run("propagates the dao error", func(t *testing.T) {
		impl := NewCorpSigning(&fakeDAO{
			getDocsFn: func(filter, project bson.M, result interface{}) error {
				return errors.New("db error")
			},
		})

		if _, err := impl.FindCorpSummary("link1", "example.com"); err == nil {
			t.Error("FindCorpSummary should propagate the dao error")
		}
	})
}
