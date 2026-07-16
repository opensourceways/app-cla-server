package watch

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func newTestImpl() *notifyAdminWatchImpl {
	remindDays := 90
	return &notifyAdminWatchImpl{
		config: &NotifyAdminConfig{
			NotifyCorpAdminRemindDays:  &remindDays,
			NotifyIndividualRemindDays: &remindDays,
			NotifyBatchSize:            500,
		},
		defaultGracePeriodDays: 30,
	}
}

func TestHandleCorpSigningNoPDF(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{Id: "link1"}
	corp := &repository.CorpSigningSummary{HasPDF: false}

	impl.handleCorpSigning(link, corp)
}

func TestHandleCorpSigningAlreadyLatest(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id: "link1",
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		},
	}
	corp := &repository.CorpSigningSummary{
		HasPDF: true,
		Link:   domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "10", Language: dp.CreateLanguage("en")}},
	}

	impl.handleCorpSigning(link, corp)
}

func TestHandleCorpSigningNoLatestCorpCLA(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id:   "link1",
		Clas: []domain.CLA{},
	}
	corp := &repository.CorpSigningSummary{
		HasPDF: true,
		Link:   domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}},
	}

	impl.handleCorpSigning(link, corp)
}

func TestHandleCorpSigningNotifyWithin7Days(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id: "link1",
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		},
	}

	corp := &repository.CorpSigningSummary{
		HasPDF:    true,
		CLANotify: "10",
		Link:      domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}},
	}

	impl.handleCorpSigning(link, corp)
}

func TestHandleIndividualSigningAlreadyLatest(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id: "link1",
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
		},
	}

	is := &domain.IndividualSigning{
		Link: domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "10", Language: dp.CreateLanguage("en")}},
	}

	impl.handleIndividualSigning(link, is)
}

func TestHandleIndividualSigningNoLatestCLA(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id:   "link1",
		Clas: []domain.CLA{},
	}

	is := &domain.IndividualSigning{
		Link: domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}},
	}

	impl.handleIndividualSigning(link, is)
}
