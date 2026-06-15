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

func NotifyAdminWatchStart(cfg *NotifyAdminConfig, lk repoLink, corp corpSigningRepo, claPlatformURL string) {
	notifyAdminWatchInstance = &notifyAdminWatchImpl{
		config:          cfg,
		link:            lk,
		corpSigningRepo: corp,
		claPlatformURL:  claPlatformURL,
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

type notifyAdminWatchImpl struct {
	config *NotifyAdminConfig

	link            repoLink
	corpSigningRepo corpSigningRepo
	claPlatformURL  string

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
	}
}

func (impl *notifyAdminWatchImpl) handleCorpSigning(link *repository.LinkCLA, corp *repository.CorpSigningSummary) {
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
	builder := emailtmpl.CLAUpdated{
		Org:              link.Org.Alias,
		CorpName:         corp.Corp.Name.CorpName(),
		AdminName:        corp.Admin.Name.Name(),
		UpdateDate:       time.Now().Format("2006-01-02"),
		ProjectURL:       link.Org.ProjectURL,
		URLOfCLAPlatform: impl.claPlatformURL + link.Id,
	}
	emailMsg, err := builder.GenEmailMsg()
	if err != nil {
		return err
	}

	emailMsg.From = link.Email.Addr.EmailAddr()
	emailMsg.To = []string{corp.Admin.EmailAddr.EmailAddr()}
	emailMsg.Subject = fmt.Sprintf("%s CLA 协议已更新 - 无需立即操作", link.Org.Alias)

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
