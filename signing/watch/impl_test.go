package watch

import (
	"errors"
	"reflect"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
)

// fakeCorpSigning records the calls made by handle.
type fakeCorpSigning struct {
	findErr          error
	findErrOnReload  error // returned by the second and later Find calls
	addEmployeeErr   error
	resetErr         error
	findVersions     []int
	currentFind      int
	resetVersions    []string
	addedEmployees   []*domain.EmployeeSigning
	findCalls        int
	addEmployeeCall  int
	resetCalls       int
	lastResetVersion int
}

func (f *fakeCorpSigning) corp(cs *domain.CorpSigning) domain.CorpSigning {
	version := 0
	if f.currentFind < len(f.findVersions) {
		version = f.findVersions[f.currentFind]
	}
	f.currentFind++
	f.findCalls++

	cs1 := *cs
	cs1.Version = version

	return cs1
}

func (f *fakeCorpSigning) ListTriggered() ([]repositoryimpl.TriggeredCorp, error) {
	return nil, nil
}

func (f *fakeCorpSigning) ResetTriggered(csId string, version int) error {
	f.resetCalls++
	f.resetVersions = append(f.resetVersions, csId)
	f.lastResetVersion = version
	return f.resetErr
}

func (f *fakeCorpSigning) Find(index string) (domain.CorpSigning, error) {
	if f.findErr != nil {
		return domain.CorpSigning{}, f.findErr
	}
	if f.findCalls >= 1 && f.findErrOnReload != nil {
		return domain.CorpSigning{}, f.findErrOnReload
	}

	cs := &domain.CorpSigning{
		Id:   index,
		Date: "2026-01-02",
		Link: domain.LinkInfo{
			Id: "link1",
			CLAInfo: domain.CLAInfo{
				CLAId:    "cla1",
				Language: dp.CreateLanguage("en"),
			},
		},
		Corp: domain.Corporation{
			PrimaryEmailDomain: "uni.edu.cn",
			AllEmailDomains:    []string{"uni.edu.cn"},
		},
	}

	return f.corp(cs), nil
}

func (f *fakeCorpSigning) AddEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	f.addEmployeeCall++

	es1 := *es
	f.addedEmployees = append(f.addedEmployees, &es1)

	return f.addEmployeeErr
}

// fakeIndividualSigning records the calls made by handle.
type fakeIndividualSigning struct {
	findByDomainsErr error
	removeAllErr     error
	records          []domain.IndividualSigning
	removeAllCalls   int
	findCalls        int
	removedDomains   [][]string
}

func (f *fakeIndividualSigning) RemoveAll(linkId string, domains []string) error {
	f.removeAllCalls++
	f.removedDomains = append(f.removedDomains, domains)
	return f.removeAllErr
}

func (f *fakeIndividualSigning) FindByDomains(linkId string, domains []string) ([]domain.IndividualSigning, error) {
	f.findCalls++
	if f.findByDomainsErr != nil {
		return nil, f.findByDomainsErr
	}

	return f.records, nil
}

func newIndividual(email, date string) domain.IndividualSigning {
	return domain.IndividualSigning{
		Date: date,
		Rep: domain.Representative{
			Name:      dp.CreateName("stu"),
			EmailAddr: dp.CreateEmailAddr(email),
		},
		AllInfo: domain.AllSingingInfo{"name": "stu"},
	}
}

func newImpl(cs corpSigning, ins individualSigning) *watchingImpl {
	return &watchingImpl{cs: cs, ins: ins}
}

func triggered() repositoryimpl.TriggeredCorp {
	return repositoryimpl.TriggeredCorp{
		Id:      "csid1",
		LinkId:  "link1",
		Domains: []string{"uni.edu.cn"},
	}
}

func TestHandleCollectsIndividualSignings(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3, 4}}
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{
			newIndividual("stu1@uni.edu.cn", "2026-01-01"),
			newIndividual("stu2@UNI.edu.cn", "2026-01-02"), // same day as corp signing
			newIndividual("stu3@uni.edu.cn", "2026-01-03"), // later than corp signing
		},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.findCalls != 2 {
		t.Errorf("find calls = %d, want 2", cs.findCalls)
	}
	if cs.addEmployeeCall != 2 {
		t.Errorf("added employees = %d, want 2 (same-day counts, later one skipped)", cs.addEmployeeCall)
	}
	if ins.removeAllCalls != 1 {
		t.Errorf("remove calls = %d, want 1", ins.removeAllCalls)
	}
	if cs.resetCalls != 1 {
		t.Errorf("reset calls = %d, want 1", cs.resetCalls)
	}
	if len(cs.resetVersions) != 1 || cs.resetVersions[0] != "csid1" {
		t.Errorf("reset corp ids = %v, want [csid1]", cs.resetVersions)
	}

	for i, es := range cs.addedEmployees {
		if es.Enabled {
			t.Errorf("employee %d should be disabled", i)
		}
		if es.Source != domain.EmployeeSigningSourceIndividual {
			t.Errorf("employee %d source = %q, want individual", i, es.Source)
		}
		if es.CLA.CLAId != "cla1" {
			t.Errorf("employee %d cla = %q, want the corp cla cla1", i, es.CLA.CLAId)
		}
		if len(es.Logs) != 1 || es.Logs[0].Action != "collect" {
			t.Errorf("employee %d logs = %v, want one collect entry", i, es.Logs)
		}
	}

	// collected records keep the original signing dates
	if cs.addedEmployees[0].Date != "2026-01-01" || cs.addedEmployees[1].Date != "2026-01-02" {
		t.Errorf("collected dates = %s, %s, want the original ones",
			cs.addedEmployees[0].Date, cs.addedEmployees[1].Date)
	}

	// domain matching is case insensitive
	if cs.addedEmployees[1].Rep.EmailAddr.EmailAddr() != "stu2@UNI.edu.cn" {
		t.Errorf("employee 1 email = %s, want stu2@UNI.edu.cn", cs.addedEmployees[1].Rep.EmailAddr.EmailAddr())
	}
}

func TestHandleRemovesAfterCollecting(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3, 4}}
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-01")},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.addEmployeeCall == 0 || ins.removeAllCalls == 0 || cs.resetCalls == 0 {
		t.Fatalf("calls: add=%d remove=%d reset=%d, want all non-zero",
			cs.addEmployeeCall, ins.removeAllCalls, cs.resetCalls)
	}
}

func TestHandleResetsWithReloadedVersion(t *testing.T) {
	// AddEmployee bumps the doc version, so handle must reload the corp
	// signing and reset the trigger with the fresh version (7), not the
	// stale one (3).
	cs := &fakeCorpSigning{findVersions: []int{3, 7}}
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-01")},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.lastResetVersion != 7 {
		t.Errorf("reset version = %d, want 7", cs.lastResetVersion)
	}
}

func TestHandleSkipsDuplicatedEmployee(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3, 3}}

	// The fake Find returns a corp without employees. The first record is
	// appended in memory by CollectEmployee, so the second one with the same
	// email must be skipped by the isMe dedup of the same in-memory corp.
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{
			newIndividual("stu1@uni.edu.cn", "2026-01-01"),
			newIndividual("stu1@uni.edu.cn", "2026-01-01"),
		},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.addEmployeeCall != 1 {
		t.Errorf("added employees = %d, want 1 (duplicate skipped)", cs.addEmployeeCall)
	}
	if ins.removeAllCalls != 1 || cs.resetCalls != 1 {
		t.Errorf("remove=%d reset=%d, want 1/1", ins.removeAllCalls, cs.resetCalls)
	}
}

func TestHandleAbortsWhenAddEmployeeFails(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3}, addEmployeeErr: errors.New("version conflict")}
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-01")},
	}

	newImpl(cs, ins).handle(triggered())

	if ins.removeAllCalls != 0 {
		t.Error("individual signings must not be removed when collecting fails")
	}
	if cs.resetCalls != 0 {
		t.Error("trigger must not be reset when collecting fails")
	}
}

func TestHandleAbortsWhenFindCorpFails(t *testing.T) {
	cs := &fakeCorpSigning{findErr: errors.New("db error")}
	ins := &fakeIndividualSigning{}

	newImpl(cs, ins).handle(triggered())

	if ins.findCalls != 0 || ins.removeAllCalls != 0 || cs.resetCalls != 0 {
		t.Error("nothing should happen when the corp signing cannot be loaded")
	}
}

func TestHandleAbortsWhenFindByDomainsFails(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3}}
	ins := &fakeIndividualSigning{findByDomainsErr: errors.New("db error")}

	newImpl(cs, ins).handle(triggered())

	if cs.addEmployeeCall != 0 || ins.removeAllCalls != 0 || cs.resetCalls != 0 {
		t.Error("nothing should happen when individual signings cannot be listed")
	}
}

func TestHandleAbortsWhenRemoveAllFails(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3, 4}}
	ins := &fakeIndividualSigning{
		removeAllErr: errors.New("db error"),
		records:      []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-01")},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.resetCalls != 0 {
		t.Error("trigger must not be reset when removing individual signings fails")
	}
}

func TestHandleAbortsWhenReloadCorpFails(t *testing.T) {
	cs := &fakeCorpSigning{}
	cs.findVersions = []int{3}
	cs.findErrOnReload = errors.New("db error")
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-01")},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.resetCalls != 0 {
		t.Error("trigger must not be reset when the corp signing cannot be reloaded")
	}
}

func TestHandleAbortsWhenResetTriggeredFails(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3, 4}, resetErr: errors.New("db error")}
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-01")},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.resetCalls != 1 {
		t.Errorf("reset calls = %d, want 1 (failure is logged, next round retries)", cs.resetCalls)
	}
}

func TestHandleStillRemovesLaterSignings(t *testing.T) {
	cs := &fakeCorpSigning{findVersions: []int{3, 3}}
	ins := &fakeIndividualSigning{
		records: []domain.IndividualSigning{newIndividual("stu1@uni.edu.cn", "2026-01-03")},
	}

	newImpl(cs, ins).handle(triggered())

	if cs.addEmployeeCall != 0 {
		t.Error("an individual signing later than the corp signing must not be collected")
	}
	// but the existing soft-delete behaviour still applies to it
	if ins.removeAllCalls != 1 {
		t.Error("individual signings of the corp domains must still be removed")
	}
	if !reflect.DeepEqual(ins.removedDomains[0], []string{"uni.edu.cn"}) {
		t.Errorf("removed domains = %v, want [uni.edu.cn]", ins.removedDomains[0])
	}
	if cs.resetCalls != 1 {
		t.Error("trigger should be reset when nothing needs collecting")
	}
}
