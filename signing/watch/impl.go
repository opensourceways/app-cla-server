package watch

import (
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
)

var impl *watchingImpl

func Start(
	cfg *Config,
	cs corpSigning,
	individual individualSigning,
) {
	impl = &watchingImpl{
		cs:       cs,
		ins:      individual,
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
		interval: cfg.intervalDuration(),
	}

	impl.start()

	logs.Info("start to watch corp signing")
}

func Stop() {
	if impl != nil {
		impl.exit()

		logs.Info("stop watching corp signing")
	}
}

// corpSigning
type corpSigning interface {
	ListTriggered() ([]repositoryimpl.TriggeredCorp, error)
	ResetTriggered(csId string, version int) error
}

// individualSigning
type individualSigning interface {
	RemoveAll(linkId string, domains []string) error
}

// watchingImpl
type watchingImpl struct {
	cs       corpSigning
	ins      individualSigning
	stop     chan struct{}
	stopped  chan struct{}
	interval time.Duration
}

func (impl *watchingImpl) start() {
	go impl.watch()
}

func (impl *watchingImpl) exit() {
	close(impl.stop)

	<-impl.stopped
}

func (impl *watchingImpl) watch() {
	needStop := func() bool {
		select {
		case <-impl.stop:
			return true
		default:
			return false
		}
	}

	var timer *time.Timer

	defer func() {
		if timer != nil {
			timer.Stop()
		}

		close(impl.stopped)
	}()

	for {
		triggered, err := impl.cs.ListTriggered()
		if err != nil {
			logs.Error("failed to list triggered corp signings, err: %s", err.Error())
		}

		for _, pr := range triggered {
			impl.handle(pr)

			if needStop() {
				return
			}
		}

		// time starts.
		if timer == nil {
			timer = time.NewTimer(impl.interval)
		} else {
			timer.Reset(impl.interval)
		}

		select {
		case <-impl.stop:
			return

		case <-timer.C:
		}
	}
}

func (impl *watchingImpl) handle(corp repositoryimpl.TriggeredCorp) {
	if err := impl.ins.RemoveAll(corp.LinkId, corp.Domains); err != nil {
		logs.Error(
			"failed to remove individual signings, csid:%s, err:%s",
			corp.Id, err.Error(),
		)

		return
	}

	if err := impl.cs.ResetTriggered(corp.Id, corp.Version); err != nil {
		logs.Error(
			"failed to reset triggered corp signing, csid:%s, err:%s",
			corp.Id, err.Error(),
		)
	}
}
