package watch

import (
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/worker"
)

var notifyAdminWatchInstance *notifyAdminWatchImpl

func NotifyAdminWatchStart(cfg *NotifyAdminConfig, lk repoLink, corp corpSigningRepo, individual individualSigningRepo, claPlatformURL string, defaultGracePeriodDays int) {
	notifyAdminWatchInstance = &notifyAdminWatchImpl{
		config:                 cfg,
		link:                   lk,
		corpSigningRepo:        corp,
		individualSigningRepo:  individual,
		claPlatformURL:         claPlatformURL,
		defaultGracePeriodDays: defaultGracePeriodDays,
	}

	notifyAdminWatchInstance.start()
}

func NotifyAdminWatchStop() {
	if notifyAdminWatchInstance != nil {
		notifyAdminWatchInstance.exit()

		logs.Info("stop watching send email")
	}
}

type corpSigningRepo interface {
	FindAll(linkId string) ([]repository.CorpSigningSummary, error)
	UpdateCLANotify(summary *repository.CorpSigningSummary) error
}

type individualSigningRepo interface {
	FindAllWithPagination(linkId string, offset, limit int) ([]domain.IndividualSigning, error)
	UpdateCLANotify(linkId, email, claId string, count int, notifyTime int64) error
}

type notifyAdminWatchImpl struct {
	config *NotifyAdminConfig

	link                   repoLink
	corpSigningRepo        corpSigningRepo
	individualSigningRepo  individualSigningRepo
	claPlatformURL         string
	defaultGracePeriodDays int

	wg   sync.WaitGroup
	stop chan struct{}
}

func (impl *notifyAdminWatchImpl) start() {
	impl.wg.Add(1)
	go impl.notifyCorpAdmin()
}

func (impl *notifyAdminWatchImpl) exit() {
	close(impl.stop)

	impl.wg.Wait()
}

func (impl *notifyAdminWatchImpl) notifyCorpAdmin() {
	interval := impl.config.genNotifyCorpAdminInterval()
	timer := time.NewTimer(interval)
	for {
		select {
		case <-impl.stop:
			timer.Stop()
			impl.wg.Done()
			return
		case <-timer.C:
			impl.handleNotifyJob()
			timer.Reset(interval)
		}
	}
}

func (impl *notifyAdminWatchImpl) handleNotifyJob() {
	needStop := func() bool {
		select {
		case <-impl.stop:
			return true
		default:
			return false
		}
	}

	links, err := impl.link.ListAll()
	if err != nil {
		logs.Error("list all link failed in notify job: ", err)
		return
	}

	for i := range links {
		if needStop() {
			return
		}

		link := &links[i]
		corpsSummary, err := impl.corpSigningRepo.FindAll(link.Id)
		if err != nil {
			logs.Error("list corp signing failed in notify job:", link.Id, err)
			continue
		}

		for j := range corpsSummary {
			if needStop() {
				return
			}

			impl.handleCorpSigning(link, &corpsSummary[j])
		}

		impl.handleIndividualSignings(link, needStop)
	}
}

func (impl *notifyAdminWatchImpl) handleCorpSigning(link *repository.LinkCLA, corp *repository.CorpSigningSummary) {
	if !corp.HasPDF {
		return
	}

	if impl.isCorpSigningLatest(link.Clas, corp.Link.CLAInfo) {
		return
	}

	latestClaId := impl.getLatestCorpClaId(link, corp.Link.Language)
	if latestClaId == "" {
		return
	}

	if corp.CLANotify == latestClaId {
		if time.Since(time.Unix(corp.ClaNotifyTime, 0)) < 7*24*time.Hour {
			return
		}
	}

	if err := impl.handleSendEmail(link, corp); err != nil {
		logs.Error("send cla notify email failed:", corp.Id, err)
		return
	}

	corp.CLANotify = latestClaId
	corp.ClaNotifyCount += 1
	corp.ClaNotifyTime = time.Now().Unix()
	if err := impl.corpSigningRepo.UpdateCLANotify(corp); err != nil {
		logs.Error("update cla notify failed: ", corp.Id, err)
	}
}

func (impl *notifyAdminWatchImpl) isCorpSigningLatest(latestCLAs []domain.CLA, signedInfo domain.CLAInfo) bool {
	var matchedCLA *domain.CLA

	for i := range latestCLAs {
		if latestCLAs[i].Type == dp.CLATypeCorp &&
			latestCLAs[i].Language == signedInfo.Language {
			matchedCLA = &latestCLAs[i]
			break
		}
	}

	return matchedCLA != nil && matchedCLA.Id == signedInfo.CLAId
}

func (impl *notifyAdminWatchImpl) handleSendEmail(link *repository.LinkCLA, corp *repository.CorpSigningSummary) error {
	if corp.Admin.Name == nil {
		return fmt.Errorf("failed to send email msg: admin name is null: %s", link.Id)
	}
	graceDays := impl.getEffectiveGracePeriodDays(link)
	builder := emailtmpl.CLAUpdated{
		Org:              link.Org.Alias,
		CorpName:         corp.Corp.Name.CorpName(),
		AdminName:        corp.Admin.Name.Name(),
		UpdateDate:       time.Now().Format("2006-01-02"),
		ProjectURL:       link.Org.ProjectURL,
		URLOfCLAPlatform: impl.claPlatformURL + link.Id,
		GracePeriodDays:  graceDays,
	}
	emailMsg, err := builder.GenEmailMsg()
	if err != nil {
		return err
	}

	emailMsg.From = link.Email.Addr.EmailAddr()
	emailMsg.To = []string{corp.Admin.EmailAddr.EmailAddr()}

	if corp.Link.Language.Language() == "en" {
		emailMsg.Subject = fmt.Sprintf("%s CLA Has Been Updated - No Immediate Action Required", link.Org.Alias)
	} else {
		emailMsg.Subject = fmt.Sprintf("%s CLA 协议已更新 - 无需立即操作", link.Org.Alias)
	}

	worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

	time.Sleep(impl.config.genSendEmailInterval())

	return nil
}

func (impl *notifyAdminWatchImpl) getLatestCorpClaId(link *repository.LinkCLA, language dp.Language) string {
	for i := range link.Clas {
		if link.Clas[i].Type == dp.CLATypeCorp && link.Clas[i].Language == language {
			return link.Clas[i].Id
		}
	}
	return ""
}

func (impl *notifyAdminWatchImpl) handleIndividualSignings(link *repository.LinkCLA, needStop func() bool) {
	individuals, err := impl.individualSigningRepo.FindAll(link.Id)
	if err != nil {
		logs.Error("list individual signing failed in notify job:", link.Id, err)
		return
	}

	for i := range individuals {
		if needStop() {
			return
		}

		impl.handleIndividualSigning(link, &individuals[i])
	}
}

func (impl *notifyAdminWatchImpl) handleIndividualSigning(link *repository.LinkCLA, is *domain.IndividualSigning) {
	latestClaId := impl.getLatestIndividualClaId(link, is.Link.Language)
	if latestClaId == "" {
		return
	}

	if is.HasSignedCLA(latestClaId) {
		return
	}

	if is.ClaNotify == latestClaId {
		if time.Since(time.Unix(is.ClaNotifyTime, 0)) < 7*24*time.Hour {
			return
		}
	}

	if err := impl.handleSendIndividualEmail(link, is); err != nil {
		logs.Error("send individual cla notify email failed:", is.Rep.EmailAddr.EmailAddr(), err)
		return
	}

	is.ClaNotify = latestClaId
	is.ClaNotifyCount += 1
	is.ClaNotifyTime = time.Now().Unix()
	if err := impl.individualSigningRepo.UpdateCLANotify(
		link.Id, is.Rep.EmailAddr.EmailAddr(), is.ClaNotify, is.ClaNotifyCount, is.ClaNotifyTime,
	); err != nil {
		logs.Error("update individual cla notify failed: ", is.Rep.EmailAddr.EmailAddr(), err)
	}
}

func (impl *notifyAdminWatchImpl) handleSendIndividualEmail(link *repository.LinkCLA, is *domain.IndividualSigning) error {
	if is.Rep.Name == nil {
		return fmt.Errorf("failed to send email msg: individual name is null: %s", link.Id)
	}
	graceDays := impl.getEffectiveGracePeriodDays(link)
	builder := emailtmpl.IndividualCLAUpdated{
		Org:              link.Org.Alias,
		Name:             is.Rep.Name.Name(),
		UpdateDate:       time.Now().Format("2006-01-02"),
		ProjectURL:       link.Org.ProjectURL,
		URLOfCLAPlatform: impl.claPlatformURL + link.Id,
		GracePeriodDays:  graceDays,
	}
	emailMsg, err := builder.GenEmailMsg()
	if err != nil {
		return err
	}

	emailMsg.From = link.Email.Addr.EmailAddr()
	emailMsg.To = []string{is.Rep.EmailAddr.EmailAddr()}

	if is.Link.Language.Language() == "en" {
		emailMsg.Subject = fmt.Sprintf("%s CLA Has Been Updated - No Immediate Action Required", link.Org.Alias)
	} else {
		emailMsg.Subject = fmt.Sprintf("%s CLA 协议已更新 - 无需立即操作", link.Org.Alias)
	}

	worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

	time.Sleep(impl.config.genSendEmailInterval())

	return nil
}

func (impl *notifyAdminWatchImpl) getEffectiveGracePeriodDays(link *repository.LinkCLA) int {
	if link.GracePeriodDays >= 0 {
		return link.GracePeriodDays
	}
	return impl.defaultGracePeriodDays
}

func (impl *notifyAdminWatchImpl) getLatestIndividualClaId(link *repository.LinkCLA, language dp.Language) string {
	for i := range link.Clas {
		if link.Clas[i].Type == dp.CLATypeIndividual && link.Clas[i].Language == language {
			return link.Clas[i].Id
		}
	}
	return ""
}
