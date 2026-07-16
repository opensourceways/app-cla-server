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
		config:                 cfg,
		link:                   lk,
		corpSigningRepo:        corp,
		individualRepo:         individual,
		claPlatformURL:         claPlatformURL,
		defaultGracePeriodDays: defaultGracePeriodDays,
		stop:                   make(chan struct{}),
		corpTrigger:            make(chan struct{}, 1),
		individualTrigger:      make(chan struct{}, 1),
	}

	logs.Info("notify admin watch config: interval_corp=%ds interval_ind=%ds remind_corp=%dd remind_ind=%dd batch=%d communities=%v",
		int(cfg.genNotifyCorpAdminInterval().Seconds()),
		int(cfg.genNotifyIndividualInterval().Seconds()),
		cfg.genNotifyCorpAdminRemindDays(),
		cfg.genNotifyIndividualRemindDays(),
		cfg.genNotifyBatchSize(),
		cfg.EnabledCommunityOrgs)

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

	wg                sync.WaitGroup
	stop              chan struct{}
	corpTrigger       chan struct{}
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
	interval := impl.config.genNotifyCorpAdminInterval()
	timer := time.NewTimer(impl.initialDelay(interval))
	for {
		select {
		case <-impl.stop:
			timer.Stop()
			impl.wg.Done()
			return
		case <-timer.C:
			impl.handleNotifyJob()
			timer.Reset(impl.nextDelay(interval))
		case <-impl.corpTrigger:
			impl.handleNotifyJob()
			timer.Reset(impl.nextDelay(interval))
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

	sendCount := 0
	batchSize := impl.config.genNotifyBatchSize()

	links, err := impl.link.ListAll()
	if err != nil {
		logs.Error("list all link failed in notify job: ", err)
		return
	}

	logs.Debug("corp notify job start, links=%d, batch=%d", len(links), batchSize)

	for i := range links {
		if needStop() {
			return
		}

		link := &links[i]

		if !impl.config.isCommunityEnabled(link.Org.Alias) {
			continue
		}

		corpsSummary, err := impl.corpSigningRepo.FindAll(link.Id)
		if err != nil {
			logs.Error("list corp signing failed in notify job:", link.Id, err)
			continue
		}

		logs.Debug("corp notify: link=%s org=%s, signings=%d", link.Id, link.Org.Alias, len(corpsSummary))

		for j := range corpsSummary {
			if needStop() {
				return
			}
			if sendCount >= batchSize {
				logs.Info("corp notify job reached batch limit: %d", batchSize)
				return
			}

			if impl.handleCorpSigning(link, &corpsSummary[j]) {
				sendCount++
			}
		}
	}

	logs.Debug("corp notify job done, sent=%d", sendCount)
}

func (impl *notifyAdminWatchImpl) handleCorpSigning(link *repository.LinkCLA, corp *repository.CorpSigningSummary) bool {
	if impl.isCorpSigningLatest(link.Clas, corp.Link.CLAInfo) {
		return false
	}

	latestCLA := impl.getLatestCorpCLA(link.Clas, corp.Link.Language)
	if latestCLA == nil {
		logs.Debug("corp notify skip: no matching CLA, link=%s corp=%s lang=%s", link.Id, corp.Id, corp.Link.Language.Language())
		return false
	}

	remindDays := impl.config.genNotifyCorpAdminRemindDays()
	nowUnix := time.Now().Unix()

	if corp.CLANotify != latestCLA.Id {
		corp.CLANotify = latestCLA.Id
		corp.ClaNotifyCount = 0
		corp.ClaNotifyTime = 0
		logs.Debug("corp notify reset: corp=%s old_notify=%s new_notify=%s", corp.Id, corp.CLANotify, latestCLA.Id)
	}

	daysSinceLastNotify := 0
	if corp.ClaNotifyTime > 0 {
		daysSinceLastNotify = int((nowUnix - corp.ClaNotifyTime) / 86400)
	}

	if daysSinceLastNotify < remindDays && corp.ClaNotifyCount > 0 {
		logs.Debug("corp notify skip: in cool-down, corp=%s count=%d days=%d remind=%d", corp.Id, corp.ClaNotifyCount, daysSinceLastNotify, remindDays)
		return false
	}

	if err := impl.handleSendEmail(link, corp); err != nil {
		logs.Error("send cla notify email failed:", corp.Id, err)
		return false
	}

	corp.ClaNotifyCount++
	corp.ClaNotifyTime = nowUnix
	if err := impl.corpSigningRepo.UpdateCLANotify(corp); err != nil {
		logs.Error("update cla notify failed: ", corp.Id, err)
	}

	return true
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
	if to := impl.config.genNotifyEmailTo(); to != "" {
		emailMsg.To = []string{to}
	}
	emailMsg.Subject = "CLA has been updated"

	worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

	// Sending email is done in goroutine.
	// Prevent the concurrency from being too high, which would cause the email server refused to serve.
	time.Sleep(impl.config.genSendEmailInterval())

	return nil
}

func (impl *notifyAdminWatchImpl) notifyIndividualSigner() {
	interval := impl.config.genNotifyIndividualInterval()
	timer := time.NewTimer(impl.initialDelay(interval))
	for {
		select {
		case <-impl.stop:
			timer.Stop()
			impl.wg.Done()
			return
		case <-timer.C:
			impl.handleIndividualNotifyJob()
			timer.Reset(impl.nextDelay(interval))
		case <-impl.individualTrigger:
			impl.handleIndividualNotifyJob()
			timer.Reset(impl.nextDelay(interval))
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

	sendCount := 0
	batchSize := impl.config.genNotifyBatchSize()

	links, err := impl.link.ListAll()
	if err != nil {
		logs.Error("list all link failed in individual notify job: ", err)
		return
	}

	logs.Debug("individual notify job start, links=%d, batch=%d", len(links), batchSize)

	for i := range links {
		if needStop() {
			return
		}

		link := &links[i]

		if !impl.config.isCommunityEnabled(link.Org.Alias) {
			continue
		}

		individuals, err := impl.individualRepo.FindAll(link.Id)
		if err != nil {
			logs.Error("list individual signing failed in notify job:", link.Id, err)
			continue
		}

		logs.Debug("individual notify: link=%s org=%s, signings=%d", link.Id, link.Org.Alias, len(individuals))

		for j := range individuals {
			if needStop() {
				return
			}
			if sendCount >= batchSize {
				logs.Info("individual notify job reached batch limit: %d", batchSize)
				return
			}

			if impl.handleIndividualSigning(link, &individuals[j]) {
				sendCount++
			}
		}
	}

	logs.Debug("individual notify job done, sent=%d", sendCount)
}

func (impl *notifyAdminWatchImpl) handleIndividualSigning(link *repository.LinkCLA, is *domain.IndividualSigning) bool {
	if impl.isIndividualSigningLatest(link.Clas, is.Link.CLAInfo) {
		return false
	}

	latestCLA := impl.getLatestIndividualCLA(link.Clas, is.Link.Language)
	if latestCLA == nil {
		logs.Debug("individual notify skip: no matching CLA, link=%s email=%s lang=%s", link.Id, is.Rep.EmailAddr.EmailAddr(), is.Link.Language.Language())
		return false
	}

	logs.Info("individual notify candidate: email=%s signed_cla=%s latest_cla=%s cla_notify=%s count=%d time=%d",
		is.Rep.EmailAddr.EmailAddr(), is.Link.CLAId, latestCLA.Id, is.ClaNotify, is.ClaNotifyCount, is.ClaNotifyTime)

	remindDays := impl.config.genNotifyIndividualRemindDays()
	nowUnix := time.Now().Unix()

	if is.ClaNotify != latestCLA.Id {
		is.ClaNotify = latestCLA.Id
		is.ClaNotifyCount = 0
		is.ClaNotifyTime = 0
		logs.Debug("individual notify reset: email=%s old_notify=%s new_notify=%s", is.Rep.EmailAddr.EmailAddr(), is.ClaNotify, latestCLA.Id)
	}

	daysSinceLastNotify := 0
	if is.ClaNotifyTime > 0 {
		daysSinceLastNotify = int((nowUnix - is.ClaNotifyTime) / 86400)
	}

	// 如果距上次通知不足 remindDays 天，跳过
	if daysSinceLastNotify < remindDays && is.ClaNotifyCount > 0 {
		logs.Debug("individual notify skip: in cool-down, email=%s count=%d days=%d remind=%d", is.Rep.EmailAddr.EmailAddr(), is.ClaNotifyCount, daysSinceLastNotify, remindDays)
		return false
	}

	if err := impl.handleSendIndividualEmail(link, is, latestCLA); err != nil {
		logs.Error("send individual cla notify email failed:", is.Rep.EmailAddr.EmailAddr(), err)
		return false
	}

	is.ClaNotifyCount++
	is.ClaNotifyTime = nowUnix
	if err := impl.individualRepo.UpdateCLANotify(is); err != nil {
		logs.Error("update individual cla notify failed: ", is.Rep.EmailAddr.EmailAddr(), err)
	}

	logs.Info("individual notify sent: email=%s link=%s count=%d", is.Rep.EmailAddr.EmailAddr(), link.Id, is.ClaNotifyCount)

	return true
}

func (impl *notifyAdminWatchImpl) isIndividualSigningLatest(latestCLAs []domain.CLA, signedInfo domain.CLAInfo) bool {
	var matchedCLA *domain.CLA
	for i := range latestCLAs {
		if latestCLAs[i].Type == dp.CLATypeIndividual &&
			latestCLAs[i].Language == signedInfo.Language {
			matchedCLA = &latestCLAs[i]
			break
		}
	}
	return matchedCLA != nil && matchedCLA.Id == signedInfo.CLAId
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

	// 格式化 CLA 更新时间，若 UpdatedAt 为 0（存量数据未记录）则用当前时间兜底
	claUpdatedAt := latestCLA.UpdatedAt
	if claUpdatedAt == 0 {
		claUpdatedAt = time.Now().Unix()
	}
	updateDate := time.Unix(claUpdatedAt, 0).Format("2006-01-02")

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
	if to := impl.config.genNotifyEmailTo(); to != "" {
		emailMsg.To = []string{to}
	}
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

// initialDelay 计算首次扫描的延迟。若 interval 为默认 86400（生产模式），
// 使用中午/凌晨对齐策略；否则直接使用配置的 interval（测试环境可配为短间隔）。
func (impl *notifyAdminWatchImpl) initialDelay(interval time.Duration) time.Duration {
	if interval >= 86400*time.Second {
		return nextNoonOrMidnight()
	}
	return interval
}

// nextDelay 计算后续扫描的间隔。生产模式用 12 小时（保证每日 2 次），
// 测试模式直接使用配置的 interval。
func (impl *notifyAdminWatchImpl) nextDelay(interval time.Duration) time.Duration {
	if interval >= 86400*time.Second {
		return 12 * time.Hour
	}
	return interval
}
