package messageimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/signing/watch"
)

func NewMessageImpl() *messageImpl {
	return &messageImpl{}
}

type messageImpl struct {
}

func (impl *messageImpl) CLAUpdated(msg message.CLAUpdatedMsg) {
	watch.ClaUpdatedWatchInstance().Send(msg)
}
