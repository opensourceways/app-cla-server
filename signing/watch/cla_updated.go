package watch

import (
	"os/exec"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/worker"
)

var claUpdatedWatchInstance *claUpdatedWatchImpl

func CLAUpdatedWatchStart(lc localCLA, linkRepo linkRepo, corp corpSigningRepo, py, claPlatformURL string) {
	claUpdatedWatchInstance = &claUpdatedWatchImpl{
		pythonBin:       py,
		claPlatformURL:  claPlatformURL,
		localCLA:        lc,
		linkRepo:        linkRepo,
		corpSigningRepo: corp,
		genCLADiff:      make(chan message.CLAUpdatedMsg, 10),
		sendEmail:       make(chan message.CLAUpdatedMsg, 10),
	}

	claUpdatedWatchInstance.start()
}

func ClaUpdatedWatchInstance() *claUpdatedWatchImpl {
	return claUpdatedWatchInstance
}

type localCLA interface {
	LocalPath(*domain.CLAIndex) string
	LocalPathOfDiff(index *domain.CLAIndex, signedClaId string) string
}

type corpSigningRepo interface {
	FindAll(linkId string) ([]repository.CorpSigningSummary, error)
}

type linkRepo interface {
	Find(string) (domain.Link, error)
}

type claUpdatedWatchImpl struct {
	pythonBin       string
	claPlatformURL  string
	localCLA        localCLA
	linkRepo        linkRepo
	corpSigningRepo corpSigningRepo
	genCLADiff      chan message.CLAUpdatedMsg
	sendEmail       chan message.CLAUpdatedMsg
}

func (impl *claUpdatedWatchImpl) Send(msg message.CLAUpdatedMsg) {
	impl.genCLADiff <- msg
	impl.sendEmail <- msg
}

func (impl *claUpdatedWatchImpl) start() {
	go impl.subscribeGenCLADiff()
	go impl.subscribeSendEmail()
}

func (impl *claUpdatedWatchImpl) subscribeGenCLADiff() {
	for v := range impl.genCLADiff {
		impl.handleGenCLADiff(v)
	}
}

func (impl *claUpdatedWatchImpl) subscribeSendEmail() {
	for v := range impl.sendEmail {
		impl.handleSendEmail(v)
	}
}

func (impl *claUpdatedWatchImpl) handleGenCLADiff(msg message.CLAUpdatedMsg) {
	oldPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.OldCLAId,
	})

	newPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	})

	diffFile := impl.localCLA.LocalPathOfDiff(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	}, msg.OldCLAId,
	)

	cmd := exec.Command(impl.pythonBin, "./util/generate_diff.py", oldPDFPath, newPDFPath, diffFile)
	if out, err := cmd.Output(); err != nil {
		logs.Error("gen pdf diff failed: ", diffFile, string(out), err)
	}
}

func (impl *claUpdatedWatchImpl) handleSendEmail(msg message.CLAUpdatedMsg) {
	summary, err := impl.corpSigningRepo.FindAll(msg.LinkId)
	if err != nil {
		logs.Error("list corp signing failed when cla updated:", err)

		return
	}

	link, err := impl.linkRepo.Find(msg.LinkId)
	if err != nil {
		logs.Error("get link info failed when cla updated:", err)

		return
	}

	for _, v := range summary {
		builder := emailtmpl.CLAUpdated{
			Org:              link.Org.Alias,
			AdminName:        v.Admin.Name.Name(),
			ProjectURL:       link.Org.ProjectURL,
			URLOfCLAPlatform: impl.claPlatformURL + msg.LinkId,
		}

		emailMsg, err1 := builder.GenEmailMsg()
		if err1 != nil {
			logs.Error(err1)

			continue
		}

		emailMsg.From = link.Email.Addr.EmailAddr()
		emailMsg.To = []string{v.Admin.EmailAddr.EmailAddr()}
		emailMsg.Subject = "CLA has been updated"

		worker.GetEmailWorker().SendSimpleMessage(link.Email.Platform, &emailMsg)

		// 防止并发太高，邮件服务器拒绝服务？
		time.Sleep(time.Second)
	}
}
