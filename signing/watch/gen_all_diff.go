package watch

import (
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/util"
)

type link interface {
	ListAll() ([]repository.LinkCLA, error)
}

func (impl *claUpdatedWatchImpl) genAllDiffFile() {
	for {
		select {
		case <-impl.stop:
			impl.wg.Done()
			return
		case <-time.After(impl.config.genAllDiffInterval()):
			impl.handleJob()
		}
	}
}

func (impl *claUpdatedWatchImpl) handleJob() {
	links, err := impl.link.ListAll()
	if err != nil {
		logs.Error("list all link failed: ", err)
		return
	}

	for _, v := range links {
		impl.handleLink(&v)
	}
}

func (impl *claUpdatedWatchImpl) handleLink(link *repository.LinkCLA) {
	for _, newCLA := range link.Clas {
		for _, oldCLA := range link.RemovedCLAs {
			if newCLA.Language != oldCLA.Language || newCLA.Type != oldCLA.Type {
				continue
			}

			diffFile := impl.localCLA.LocalPathOfDiff(&domain.CLAIndex{
				LinkId: link.Id,
				CLAId:  newCLA.Id,
			},
				oldCLA.Id,
			)

			if !util.IsFileNotExist(diffFile) {
				continue
			}

			impl.genCLADiffByCron <- message.CLAUpdatedMsg{
				LinkId:   link.Id,
				OldCLAId: oldCLA.Id,
				NewCLAId: newCLA.Id,
			}
		}
	}
}
