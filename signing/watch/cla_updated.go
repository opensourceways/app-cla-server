package watch

import (
	"os/exec"
	"sync"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/util"
)

var claUpdatedWatchInstance *claUpdatedWatchImpl

func CLAUpdatedWatchStart(lk link, lc localCLA, cfg *Config, py string) {
	claUpdatedWatchInstance = &claUpdatedWatchImpl{
		config:           cfg,
		link:             lk,
		localCLA:         lc,
		pythonBin:        py,
		stop:             make(chan struct{}),
		genCLADiff:       make(chan message.CLAUpdatedMsg, cfg.MsgChannelSize),
		genCLADiffByCron: make(chan message.CLAUpdatedMsg, cfg.MsgChannelSize),
	}

	claUpdatedWatchInstance.start()
}

func CLAUpdateWatchStop() {
	if claUpdatedWatchInstance != nil {
		claUpdatedWatchInstance.exit()

		logs.Info("stop watching cla updated")
	}
}

func SendCLAUpdatedEvent(msg message.CLAUpdatedMsg) {
	select {
	case claUpdatedWatchInstance.genCLADiff <- msg:
	default:
		logs.Error("send cla updated event failed: ", msg)
	}
}

type localCLA interface {
	LocalPath(*domain.CLAIndex) string
	LocalPathOfDiff(index *domain.CLAIndex, signedClaId string) string
}

type claUpdatedWatchImpl struct {
	wg sync.WaitGroup

	config           *Config
	link             link
	localCLA         localCLA
	pythonBin        string
	stop             chan struct{}
	genCLADiff       chan message.CLAUpdatedMsg
	genCLADiffByCron chan message.CLAUpdatedMsg
}

func (impl *claUpdatedWatchImpl) start() {
	impl.wg.Add(1)
	go impl.subscribeGenCLADiff()

	impl.wg.Add(1)
	go impl.genAllDiffFile()
}

func (impl *claUpdatedWatchImpl) exit() {
	close(impl.stop)

	impl.wg.Wait()
}

func (impl *claUpdatedWatchImpl) subscribeGenCLADiff() {
	for {
		select {
		case <-impl.stop:
			impl.wg.Done()
			return
		case msgPrimary := <-impl.genCLADiff:
			impl.handleGenCLADiff(msgPrimary)
		case msgSecondary := <-impl.genCLADiffByCron:
		priority:
			for {
				select {
				case msgPrimary := <-impl.genCLADiff:
					impl.handleGenCLADiff(msgPrimary)
				default:
					break priority
				}
			}

			impl.handleGenCLADiff(msgSecondary)
		}
	}
}

func (impl *claUpdatedWatchImpl) handleGenCLADiff(msg message.CLAUpdatedMsg) {
	diffFile := impl.localCLA.LocalPathOfDiff(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	},
		msg.OldCLAId,
	)

	if !util.IsFileNotExist(diffFile) {
		return
	}

	oldPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.OldCLAId,
	})

	newPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	})

	for i := 0; i < impl.config.PythonRetryTimes; i++ {
		cmd := exec.Command(impl.pythonBin, "./util/generate_diff.py", oldPDFPath, newPDFPath, diffFile)
		if out, err := cmd.Output(); err != nil {
			logs.Error("gen pdf diff failed: ", diffFile, string(out), err)
		} else {
			return
		}
	}
}
