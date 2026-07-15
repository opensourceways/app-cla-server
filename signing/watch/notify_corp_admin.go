package watch

import (
	"fmt"
	"net/url"
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
		config:                cfg,
		link:                  lk,
		corpSigningRepo:       corp,
		individualRepo:        individual,
		claPlatformURL:        claPlatformURL,
		defaultGracePeriodDays: defaultGracePeriodDays,
		stop:                   make(chan struct{}),
		corpTrigger:            make(chan struct{}, 1),
		individualTrigger:      make(chan struct{}, 1),
	}

	notifyAdminWatchInstance.start()
}

func NotifyAdminWatchStop() {
	if notifyAdminWatchInstance != nil {
		notifyAdminWatchInstance.exit()

		logs.Info("stop watching send email")
	}
}

// TriggerNotify 立即触发一次通知扫描，不等待定时器周期。
// CLA 更新后应调用此函数，确保企业和个人都在最短时间内收到通知。
func TriggerNotify() {
	if notifyAdminWatchInstance == nil {
		return
	}
	select {
	case notifyAdminWatchInstance.corpTrigger <- struct{}{}:
	default:
	}
	select {
	case notifyAdminWatchInstance.individualTrigger <- struct{}{}:
	default:
	}
}

type corpSigningRepo interface {
	FindAll(linkId string) ([]repository.CorpSigningSummary, error)
	UpdateCLANotify(summary *repository.CorpSigningSummary) error
}

type individualSigningRepo interface {
	FindAll(linkId string) ([]domain.IndividualSigning, error)
	UpdateCLANotify(is *domain.IndividualSigning) error
}

type notifyAdminWatchImpl struct {
	config *NotifyAdminConfig

	link            repoLink
	corpSigningRepo corpSigningRepo
	individualRepo  individualSigningRepo
	claPlatformURL  string

	defaultGracePeriodDays int

	wg               sync.WaitGroup
	stop             chan struct{}
	corpTrigger      chan struct{}
	individualTrigger chan struct{}
}

func (impl *notifyAdminWatchImpl) start() {
	impl.wg.Add(1)
	go impl.notifyCorpAdmin()

	impl.wg.Add(1)
	go impl.notifyIndividualSigner()
}

func (impl *notifyAdminWatchImpl) exit() {
	close(impl.stop)

	impl.wg.Wait()
}

func (impl *notifyAdminWatchImpl) notifyCorpAdmin() {
	timer := time.NewTimer(nextNoonOrMidnight())
	for {
		select {
		case <-impl.stop:
			timer.Stop()
			impl.wg.Done()
			return
		case <-timer.C:
			impl.handleNotifyJob()
			timer.Reset(12 * time.Hour)
		case <-impl.corpTrigger:
			impl.handleNotifyJob()
			timer.Reset(nextNoonOrMidnight())
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

	latestCLA := impl.getLatestCorpCLA(link.Clas, corp.Link.Language)
	if latestCLA == nil {
		return
	}

	remindDays := impl.config.genNotifyCorpAdminRemindDays()
	nowUnix := time.Now().Unix()

	if corp.CLANotify != latestCLA.Id {
		corp.CLANotify = latestCLA.Id
		corp.ClaNotifyCount = 0
		corp.ClaNotifyTime = 0
	}

	if corp.ClaNotifyTime == 0 && corp.ClaNotifyCount == 0 && corp.CLANotify == latestCLA.Id {
		corp.ClaNotifyCount = 1
		corp.ClaNotifyTime = nowUnix
		if err := impl.corpSigningRepo.UpdateCLANotify(corp); err != nil {
			logs.Error("init cla_notify_count/time failed: ", corp.Id, err)
		}
		return
	}

	daysSinceLastNotify := 0
	if corp.ClaNotifyTime > 0 {
		daysSinceLastNotify = int((nowUnix - corp.ClaNotifyTime) / 86400)
	}

	if daysSinceLastNotify < remindDays && corp.ClaNotifyCount > 0 {
		return
	}

	if err := impl.handleSendEmail(link, corp); err != nil {
		logs.Error("send cla notify email failed:", corp.Id, err)
		return
	}

	corp.ClaNotifyCount++
	corp.ClaNotifyTime = nowUnix
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
		AdminName:        corp.Admin.Name.Name(),
		ProjectURL:       link.Org.ProjectURL,
		URLOfCLAPlatform: impl.rootURL() + "/corporation-manager-login/" + link.Id,
	}
	emailMsg, err := builder.GenEmailMsg()
	if err != nil {
		return err
	}

	emailMsg.From = link.Email.Addr.EmailAddr()
	emailMsg.To = []string{corp.Admin.EmailAddr.EmailAddr()}
	emailMsg.Subject = "CLA has been updated"

	worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

	// Sending email is done in goroutine.
	// Prevent the concurrency from being too high, which would cause the email server refused to serve.
	time.Sleep(impl.config.genSendEmailInterval())

	return nil
}

func (impl *notifyAdminWatchImpl) notifyIndividualSigner() {
	timer := time.NewTimer(nextNoonOrMidnight())
	for {
		select {
		case <-impl.stop:
			timer.Stop()
			impl.wg.Done()
			return
		case <-timer.C:
			impl.handleIndividualNotifyJob()
			timer.Reset(12 * time.Hour)
		case <-impl.individualTrigger:
			impl.handleIndividualNotifyJob()
			timer.Reset(nextNoonOrMidnight())
		}
	}
}

func (impl *notifyAdminWatchImpl) handleIndividualNotifyJob() {
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
		logs.Error("list all link failed in individual notify job: ", err)
		return
	}

	for i := range links {
		if needStop() {
			return
		}

		link := &links[i]
		individuals, err := impl.individualRepo.FindAll(link.Id)
		if err != nil {
			logs.Error("list individual signing failed in notify job:", link.Id, err)
			continue
		}

		for j := range individuals {
			if needStop() {
				return
			}

			impl.handleIndividualSigning(link, &individuals[j])
		}
	}
}

func (impl *notifyAdminWatchImpl) handleIndividualSigning(link *repository.LinkCLA, is *domain.IndividualSigning) {
	// 检查是否签署了最新版本
	if impl.isIndividualSigningLatest(link.Clas, is.Link.CLAInfo) {
		return
	}

	// 获取当前 CLA 的更新时间
	latestCLA := impl.getLatestIndividualCLA(link.Clas, is.Link.Language)
	if latestCLA == nil {
		return
	}

	remindDays := impl.config.genNotifyIndividualRemindDays()
	nowUnix := time.Now().Unix()

	if is.ClaNotify != latestCLA.Id {
		is.ClaNotify = latestCLA.Id
		is.ClaNotifyCount = 0
		is.ClaNotifyTime = 0
	}

	if is.ClaNotifyTime == 0 && is.ClaNotifyCount == 0 && is.ClaNotify == latestCLA.Id {
		is.ClaNotifyCount = 1
		is.ClaNotifyTime = nowUnix
		if err := impl.individualRepo.UpdateCLANotify(is); err != nil {
			logs.Error("init cla_notify_count/time failed: ", is.Rep.EmailAddr.EmailAddr(), err)
		}
		return
	}

	// 计算距离上次通知的天数
	daysSinceLastNotify := 0
	if is.ClaNotifyTime > 0 {
		daysSinceLastNotify = int((nowUnix - is.ClaNotifyTime) / 86400)
	}

	// 如果距上次通知不足 remindDays 天，跳过
	if daysSinceLastNotify < remindDays && is.ClaNotifyCount > 0 {
		return
	}

	if err := impl.handleSendIndividualEmail(link, is, latestCLA); err != nil {
		logs.Error("send individual cla notify email failed:", is.Rep.EmailAddr.EmailAddr(), err)
		return
	}

	is.ClaNotifyCount++
	is.ClaNotifyTime = nowUnix
	if err := impl.individualRepo.UpdateCLANotify(is); err != nil {
		logs.Error("update individual cla notify failed: ", is.Rep.EmailAddr.EmailAddr(), err)
	}
}

func (impl *notifyAdminWatchImpl) isIndividualSigningLatest(latestCLAs []domain.CLA, signedInfo domain.CLAInfo) bool {
	for i := range latestCLAs {
		if latestCLAs[i].Type == dp.CLATypeIndividual &&
			latestCLAs[i].Language == signedInfo.Language {
			return latestCLAs[i].Id == signedInfo.CLAId
		}
	}
	return true
}

func (impl *notifyAdminWatchImpl) getLatestIndividualCLA(clas []domain.CLA, language dp.Language) *domain.CLA {
	for i := range clas {
		if clas[i].Type == dp.CLATypeIndividual && clas[i].Language == language {
			return &clas[i]
		}
	}
	return nil
}

func (impl *notifyAdminWatchImpl) getLatestCorpCLA(clas []domain.CLA, language dp.Language) *domain.CLA {
	for i := range clas {
		if clas[i].Type == dp.CLATypeCorp && clas[i].Language == language {
			return &clas[i]
		}
	}
	return nil
}

func (impl *notifyAdminWatchImpl) handleSendIndividualEmail(link *repository.LinkCLA, is *domain.IndividualSigning, latestCLA *domain.CLA) error {
	name := ""
	if is.Rep.Name != nil {
		name = is.Rep.Name.Name()
	}

	// 格式化 CLA 更新时间
	updateDate := time.Unix(latestCLA.UpdatedAt, 0).Format("2006-01-02")

	// 构造个人签署 URL：从 claPlatformURL 提取 scheme+host，避免 /sign/sign-cla 双前缀
	// 最终格式：https://clasign.osinfra.cn/sign-cla/{linkId}/individual-update?email=xxx
	signURL := fmt.Sprintf("%s/sign-cla/%s/individual-update?email=%s",
		impl.rootURL(), link.Id, is.Rep.EmailAddr.EmailAddr())

	builder := emailtmpl.CLAUpdatedIndividual{
		Name:            name,
		Org:             link.Org.Alias,
		UpdateDate:      updateDate,
		GracePeriodDays: link.GetEffectiveGracePeriodDays(impl.defaultGracePeriodDays),
		SignCLAURL:      signURL,
		ProjectURL:      link.Org.ProjectURL,
	}
	emailMsg, err := builder.GenEmailMsg()
	if err != nil {
		return err
	}

	emailMsg.From = link.Email.Addr.EmailAddr()
	emailMsg.To = []string{is.Rep.EmailAddr.EmailAddr()}
	emailMsg.Subject = "CLA has been updated - Action Required"

	worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

	time.Sleep(impl.config.genSendEmailInterval())

	return nil
}

// rootURL 从 claPlatformURL 中提取 scheme+host（去掉路径前缀），
// 用于构造前端签署页面 URL，避免 /sign/sign-cla 双前缀问题。
// 例如 https://clasign.osinfra.cn/sign/ -> https://clasign.osinfra.cn
func (impl *notifyAdminWatchImpl) rootURL() string {
	u, err := url.Parse(impl.claPlatformURL)
	if err != nil {
		return impl.claPlatformURL
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func nextNoonOrMidnight() time.Duration {
	now := time.Now()
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	if now.After(noon) {
		noon = noon.Add(24 * time.Hour)
	}
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	if noon.Before(midnight) {
		return noon.Sub(now)
	}
	return midnight.Sub(now)
}
